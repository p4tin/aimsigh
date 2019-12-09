package aimsigh

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func CreateDao() Discoverer {
	client, _ := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	return ServicesDao{
		servicesCollection: client.Database("DiscoveryService").Collection("Services"),
		InstanceID:         xid.New().String(),
	}
}

type ServicesDao struct {
	servicesCollection *mongo.Collection
	InstanceID         string
}

func (sd ServicesDao) UpdateAliveRecord(dc, sn, ip string, port int) (RegistrationRecord, error) {
	key := bson.M{
		"data_center_id": dc,
		"service_name":   sn,
		"instance_id":    sd.InstanceID,
	}

	registration := RegistrationRecord{
		DataCenterId: dc,
		ServiceName:  sn,
		InstanceId:   sd.InstanceID,
		CreatedAt:    time.Now(),
		IpAddress:    ip,
		Port:         port,
	}

	update := bson.M{
		"$set": registration,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := sd.servicesCollection.FindOneAndUpdate(ctx, key, update, &opt).Decode(&registration)

	if err != nil {
		return RegistrationRecord{}, err
	}

	return registration, nil
}

func (sd ServicesDao) GetServiceAddress(dataCenter, serviceName string) (string, error) {
	registrations := make([]*RegistrationRecord, 0)
	key := bson.M{
		"data_center_id": dataCenter,
		"service_name":   serviceName,
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := sd.servicesCollection.Find(ctx, key)
	if err != nil {
		return "", err
	}

	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var elem RegistrationRecord
		err := cur.Decode(&elem)
		if err != nil {
			return "", err
		}

		registrations = append(registrations, &elem)
	}
	if len(registrations) == 0 {
		return "", errors.New(fmt.Sprintf("no %s servers available in DC %s", serviceName, dataCenter))
	}
	index := rand.Intn(len(registrations))

	return fmt.Sprintf("%s:%d", registrations[index].IpAddress, registrations[index].Port), nil
}
