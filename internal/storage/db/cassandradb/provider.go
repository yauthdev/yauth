package cassandradb

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	cansandraDriver "github.com/gocql/gocql"
	"github.com/rs/zerolog"

	"github.com/authorizerdev/authorizer/internal/config"
	"github.com/authorizerdev/authorizer/internal/crypto"
	"github.com/authorizerdev/authorizer/internal/storage/schemas"
)

// Dependencies struct the cassandradb data store provider
type Dependencies struct {
	Log *zerolog.Logger
}

type provider struct {
	config       *config.Config
	dependencies *Dependencies
	db           *cansandraDriver.Session
}

// KeySpace for the cassandra database
var KeySpace string

// NewProvider to initialize arangodb connection
func NewProvider(cfg *config.Config, deps *Dependencies) (*provider, error) {
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbHost := cfg.DatabaseHost
		dbPort := cfg.DatabasePort
		if dbPort != 0 && dbHost != "" {
			dbURL = fmt.Sprintf("%s:%d", dbHost, dbPort)
		} else if dbHost != "" {
			dbURL = dbHost
		}
	}

	KeySpace = cfg.DatabaseName
	if KeySpace == "" {
		return nil, fmt.Errorf("database name is required for cassandra. It is used as keyspace in case of cassandra")
	}
	clusterURL := []string{}
	if strings.Contains(dbURL, ",") {
		clusterURL = strings.Split(dbURL, ",")
	} else {
		clusterURL = append(clusterURL, dbURL)
	}
	cassandraClient := cansandraDriver.NewCluster(clusterURL...)
	dbUsername := cfg.DatabaseUsername
	dbPassword := cfg.DatabasePassword

	if dbUsername != "" && dbPassword != "" {
		cassandraClient.Authenticator = &cansandraDriver.PasswordAuthenticator{
			Username: dbUsername,
			Password: dbPassword,
		}
	}

	dbCert := cfg.DatabaseCert
	dbCACert := cfg.DatabaseCACert
	dbCertKey := cfg.DatabaseCertKey
	if dbCert != "" && dbCACert != "" && dbCertKey != "" {
		certString, err := crypto.DecryptB64(dbCert)
		if err != nil {
			return nil, err
		}

		keyString, err := crypto.DecryptB64(dbCertKey)
		if err != nil {
			return nil, err
		}

		caString, err := crypto.DecryptB64(dbCACert)
		if err != nil {
			return nil, err
		}

		cert, err := tls.X509KeyPair([]byte(certString), []byte(keyString))
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caString))

		cassandraClient.SslOpts = &cansandraDriver.SslOptions{
			Config: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            caCertPool,
				InsecureSkipVerify: true,
			},
			EnableHostVerification: false,
		}
	}

	cassandraClient.RetryPolicy = &cansandraDriver.SimpleRetryPolicy{
		NumRetries: 3,
	}
	cassandraClient.Consistency = gocql.LocalQuorum
	cassandraClient.ConnectTimeout = 10 * time.Second
	cassandraClient.ProtoVersion = 4
	cassandraClient.Timeout = 30 * time.Minute // for large data

	session, err := cassandraClient.CreateSession()
	if err != nil {
		return nil, err
	}

	// Note for astra keyspaces can only be created from there console
	// https://docs.datastax.com/en/astra/docs/datastax-astra-faq.html#_i_am_trying_to_create_a_keyspace_in_the_cql_shell_and_i_am_running_into_an_error_how_do_i_fix_this
	getKeyspaceQuery := fmt.Sprintf("SELECT keyspace_name FROM system_schema.keyspaces;")
	scanner := session.Query(getKeyspaceQuery).Iter().Scanner()
	hasAuthorizerKeySpace := false
	for scanner.Next() {
		var keySpace string
		err := scanner.Scan(&keySpace)
		if err != nil {
			return nil, err
		}
		if keySpace == KeySpace {
			hasAuthorizerKeySpace = true
			break
		}
	}

	if !hasAuthorizerKeySpace {
		createKeySpaceQuery := fmt.Sprintf("CREATE KEYSPACE %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};", KeySpace)
		err = session.Query(createKeySpaceQuery).Exec()
		if err != nil {
			return nil, err
		}
	}

	// make sure collections are present
	envCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, env text, hash text, updated_at bigint, created_at bigint, PRIMARY KEY (id))",
		KeySpace, schemas.Collections.Env)
	err = session.Query(envCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}

	sessionCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, user_id text, user_agent text, ip text, updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.Session)
	err = session.Query(sessionCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	sessionIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_session_user_id ON %s.%s (user_id)", KeySpace, schemas.Collections.Session)
	err = session.Query(sessionIndexQuery).Exec()
	if err != nil {
		return nil, err
	}

	userCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, email text, email_verified_at bigint, password text, signup_methods text, given_name text, family_name text, middle_name text, nickname text, gender text, birthdate text, phone_number text, phone_number_verified_at bigint, picture text, roles text, updated_at bigint, created_at bigint, revoked_timestamp bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.User)
	err = session.Query(userCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	userIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_user_email ON %s.%s (email)", KeySpace, schemas.Collections.User)
	err = session.Query(userIndexQuery).Exec()
	if err != nil {
		return nil, err
	}

	userPhoneNumberIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_user_phone_number ON %s.%s (phone_number)", KeySpace, schemas.Collections.User)
	err = session.Query(userPhoneNumberIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	// add is_multi_factor_auth_enabled on users table
	userTableAlterQuery := fmt.Sprintf(`ALTER TABLE %s.%s ADD is_multi_factor_auth_enabled boolean`, KeySpace, schemas.Collections.User)
	err = session.Query(userTableAlterQuery).Exec()
	if err != nil {
		deps.Log.Debug().Err(err).Msg("Failed to alter table as is_multi_factor_auth_enabled column exists")
		// continue
	}

	// token is reserved keyword in cassandra, hence we need to use jwt_token
	verificationRequestCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, jwt_token text, identifier text, expires_at bigint, email text, nonce text, redirect_uri text, created_at bigint, updated_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.VerificationRequest)
	err = session.Query(verificationRequestCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	verificationRequestIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_verification_request_email ON %s.%s (email)", KeySpace, schemas.Collections.VerificationRequest)
	err = session.Query(verificationRequestIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	verificationRequestIndexQuery = fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_verification_request_identifier ON %s.%s (identifier)", KeySpace, schemas.Collections.VerificationRequest)
	err = session.Query(verificationRequestIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	verificationRequestIndexQuery = fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_verification_request_jwt_token ON %s.%s (jwt_token)", KeySpace, schemas.Collections.VerificationRequest)
	err = session.Query(verificationRequestIndexQuery).Exec()
	if err != nil {
		return nil, err
	}

	webhookCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, event_name text, endpoint text, enabled boolean, headers text, updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.Webhook)
	err = session.Query(webhookCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	webhookIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_webhook_event_name ON %s.%s (event_name)", KeySpace, schemas.Collections.Webhook)
	err = session.Query(webhookIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	// add event_description to webhook table
	webhookAlterQuery := fmt.Sprintf(`ALTER TABLE %s.%s ADD (event_description text);`, KeySpace, schemas.Collections.Webhook)
	err = session.Query(webhookAlterQuery).Exec()
	if err != nil {
		deps.Log.Debug().Err(err).Msg("Failed to alter table as event_description column exists")
		// continue
	}

	webhookLogCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, http_status bigint, response text, request text, webhook_id text,updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.WebhookLog)
	err = session.Query(webhookLogCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	webhookLogIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_webhook_log_webhook_id ON %s.%s (webhook_id)", KeySpace, schemas.Collections.WebhookLog)
	err = session.Query(webhookLogIndexQuery).Exec()
	if err != nil {
		return nil, err
	}

	emailTemplateCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, event_name text, template text, updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.EmailTemplate)
	err = session.Query(emailTemplateCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}
	emailTemplateIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_email_template_event_name ON %s.%s (event_name)", KeySpace, schemas.Collections.EmailTemplate)
	err = session.Query(emailTemplateIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	// add subject on email_templates table
	emailTemplateAlterQuery := fmt.Sprintf(`ALTER TABLE %s.%s ADD (subject text, design text);`, KeySpace, schemas.Collections.EmailTemplate)
	err = session.Query(emailTemplateAlterQuery).Exec()
	if err != nil {
		deps.Log.Debug().Err(err).Msg("Failed to alter table as subject & design column exists")
		// continue
	}

	otpCollection := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, email text, otp text, expires_at bigint, updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.OTP)
	err = session.Query(otpCollection).Exec()
	if err != nil {
		return nil, err
	}
	otpIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_otp_email ON %s.%s (email)", KeySpace, schemas.Collections.OTP)
	err = session.Query(otpIndexQuery).Exec()
	if err != nil {
		return nil, err
	}
	// Add phone_number column to otp table
	otpAlterQuery := fmt.Sprintf(`ALTER TABLE %s.%s ADD (phone_number text);`, KeySpace, schemas.Collections.OTP)
	err = session.Query(otpAlterQuery).Exec()
	if err != nil {
		deps.Log.Debug().Err(err).Msg("Failed to alter table as phone_number column exists")
		// continue
	}
	// Add app_data column to users table
	appDataAlterQuery := fmt.Sprintf(`ALTER TABLE %s.%s ADD (app_data text);`, KeySpace, schemas.Collections.User)
	err = session.Query(appDataAlterQuery).Exec()
	if err != nil {
		deps.Log.Debug().Err(err).Msg("Failed to alter table as app_data column exists")
		// continue
	}
	// Add phone number index
	otpIndexQueryPhoneNumber := fmt.Sprintf("CREATE INDEX IF NOT EXISTS authorizer_otp_phone_number ON %s.%s (phone_number)", KeySpace, schemas.Collections.OTP)
	err = session.Query(otpIndexQueryPhoneNumber).Exec()
	if err != nil {
		return nil, err
	}
	// add authenticators table
	totpCollectionQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (id text, user_id text, method text, secret text, recovery_codes text, verified_at bigint, updated_at bigint, created_at bigint, PRIMARY KEY (id))", KeySpace, schemas.Collections.Authenticators)
	err = session.Query(totpCollectionQuery).Exec()
	if err != nil {
		return nil, err
	}

	return &provider{
		config:       cfg,
		dependencies: deps,
		db:           session,
	}, err
}