package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	ssoOidcTypes "github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/skratchdot/open-golang/open"
)

func GenerateConfig(ssoURL, region string, overwrite bool) error {
	startUrl := ssoURL
	ctx := context.Background()
	token, err := newOIDCToken(ctx, startUrl, region)
	if err != nil {
		log.Fatalf("unable to create OIDC token: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := sso.NewFromConfig(cfg)
	accounts := ListAccounts(ctx, client, token.AccessToken)

	if overwrite {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath := homeDir + "/.aws/config"
		err = os.Remove(configPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		for _, acc := range accounts {
			roles := GetRolesByAccount(ctx, client, acc, token.AccessToken)
			err := writeConfig(*acc.AccountName, region, startUrl, configPath, roles)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		for _, acc := range accounts {
			roles := GetRolesByAccount(ctx, client, acc, token.AccessToken)
			printConfig(*acc.AccountName, region, startUrl, roles)
		}
		return nil
	}
}

func writeConfig(accountName, region, startUrl, dir string, roles []types.RoleInfo) error {
	f, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	for _, role := range roles {
		str := `[profile %s-%s]
sso_start_url=%s
sso_region=%s
sso_account_id=%s
sso_role_name=%s

`
		_, err = f.WriteString(
			fmt.Sprintf(str, strings.ToLower(accountName), strings.ToLower(*role.RoleName), startUrl, region, *role.AccountId, *role.RoleName),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func printConfig(accountName, region, startUrl string, roles []types.RoleInfo) {
	for _, role := range roles {
		str := `[profile %s-%s]
sso_start_url=%s
sso_region=%s
sso_account_id=%s
sso_role_name=%s

`
		fmt.Printf(str, strings.ToLower(accountName), strings.ToLower(*role.RoleName), startUrl, region, *role.AccountId, *role.RoleName)
	}
}

func newOIDCToken(ctx context.Context, startUrl, region string) (*ssooidc.CreateTokenOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	oidc := ssooidc.NewFromConfig(cfg)

	clientCreds, err := oidc.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("acg-cli"),
		ClientType: aws.String("public"),
	})
	log.Printf("Created new OIDC client (expires at: %s)", time.Unix(clientCreds.ClientSecretExpiresAt, 0))

	deviceCreds, err := oidc.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     clientCreds.ClientId,
		ClientSecret: clientCreds.ClientSecret,
		StartUrl:     aws.String(startUrl),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Created OIDC device code for %s (expires in: %ds)", startUrl, deviceCreds.ExpiresIn)
	log.Println("Opening SSO authorization page in browser")

	_, err = fmt.Fprintf(os.Stderr, "Opening the SSO authorization page in your default browser (use Ctrl-C to abort)\n%s\n", aws.ToString(deviceCreds.VerificationUriComplete))
	if err != nil {
		return nil, err
	}
	if err := open.Run(aws.ToString(deviceCreds.VerificationUriComplete)); err != nil {
		log.Printf("Failed to open browser: %s", err)
	}

	var slowDownDelay = 5 * time.Second
	var retryInterval = 5 * time.Second

	if i := deviceCreds.Interval; i > 0 {
		retryInterval = time.Duration(i) * time.Second
	}

	for {
		t, err := oidc.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     clientCreds.ClientId,
			ClientSecret: clientCreds.ClientSecret,
			DeviceCode:   deviceCreds.DeviceCode,
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		})
		if err != nil {
			var sde *ssoOidcTypes.SlowDownException
			if errors.As(err, &sde) {
				retryInterval += slowDownDelay
			}

			var ape *ssoOidcTypes.AuthorizationPendingException
			if errors.As(err, &ape) {
				time.Sleep(retryInterval)
				continue
			}

			return nil, err
		}

		log.Printf("Created new OIDC access token for %s (expires in: %ds)", startUrl, t.ExpiresIn)
		return t, nil
	}
}

func ListAccounts(ctx context.Context, client *sso.Client, token *string) []types.AccountInfo {
	var (
		nextToken *string
		accounts  []types.AccountInfo
	)
	for {
		acc, err := client.ListAccounts(ctx, &sso.ListAccountsInput{AccessToken: token, NextToken: nextToken, MaxResults: aws.Int32(50)})
		if err != nil {
			log.Fatalf("unable to list accounts, %v", err)
		}
		accounts = append(accounts, acc.AccountList...)
		nextToken = acc.NextToken
		if nextToken == nil {
			break
		}
	}
	return accounts
}

func GetRolesByAccount(ctx context.Context, client *sso.Client, account types.AccountInfo, token *string) []types.RoleInfo {
	roles, err := client.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
		AccessToken: token,
		AccountId:   account.AccountId,
	})
	if err != nil {
		log.Fatalf("unable to list roles, %v", err)
	}
	return roles.RoleList
}
