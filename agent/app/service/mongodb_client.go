package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongodbConnectionInfo struct {
	Address    string
	Port       uint
	InitialDB  string
	Username   string
	Password   string
	Timeout    uint
	SSL        bool
	RootCert   string
	ClientKey  string
	ClientCert string
	SkipVerify bool
}

func mongodbConnectionInfoFromCreate(req dto.DatabaseCreate) mongodbConnectionInfo {
	return mongodbConnectionInfo{
		Address:    req.Address,
		Port:       req.Port,
		InitialDB:  req.InitialDB,
		Username:   req.Username,
		Password:   req.Password,
		Timeout:    req.Timeout,
		SSL:        req.SSL,
		RootCert:   req.RootCert,
		ClientKey:  req.ClientKey,
		ClientCert: req.ClientCert,
		SkipVerify: req.SkipVerify,
	}
}

func mongodbConnectionInfoFromModel(db model.Database) mongodbConnectionInfo {
	return mongodbConnectionInfo{
		Address:    db.Address,
		Port:       db.Port,
		InitialDB:  db.InitialDB,
		Username:   db.Username,
		Password:   db.Password,
		Timeout:    db.Timeout,
		SSL:        db.SSL,
		RootCert:   db.RootCert,
		ClientKey:  db.ClientKey,
		ClientCert: db.ClientCert,
		SkipVerify: db.SkipVerify,
	}
}

func loadRemoteMongodbConnection(database string) (mongodbConnectionInfo, error) {
	db, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return mongodbConnectionInfo{}, err
	}
	return mongodbConnectionInfoFromModel(db), nil
}

func newRemoteMongodbClient(info mongodbConnectionInfo) (*mongo.Client, context.Context, context.CancelFunc, error) {
	timeout := time.Duration(info.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	clientOptions := options.Client().ApplyURI(buildRemoteMongodbURI(info)).
		SetServerSelectionTimeout(timeout).
		SetConnectTimeout(timeout)
	if info.SSL {
		tlsConfig, err := buildRemoteMongodbTLSConfig(info)
		if err != nil {
			return nil, nil, nil, err
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		cancel()
		return nil, nil, nil, err
	}
	return client, ctx, cancel, nil
}

func buildRemoteMongodbURI(info mongodbConnectionInfo) string {
	uri := url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(info.Address, strconv.Itoa(int(info.Port))),
		Path:   "/",
	}
	uri.User = url.UserPassword(info.Username, info.Password)
	query := url.Values{}
	authSource := info.InitialDB
	if authSource == "" {
		authSource = "admin"
	}
	query.Set("authSource", authSource)
	query.Set("directConnection", "true")
	uri.RawQuery = query.Encode()
	return uri.String()
}

func buildRemoteMongodbTLSConfig(info mongodbConnectionInfo) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: info.SkipVerify,
	}

	if info.RootCert != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(info.RootCert)) {
			return nil, fmt.Errorf("load mongodb ca cert failed")
		}
		tlsConfig.RootCAs = pool
	}

	if info.ClientCert != "" && info.ClientKey != "" {
		cert, err := tls.X509KeyPair([]byte(info.ClientCert), []byte(info.ClientKey))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
