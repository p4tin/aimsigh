package aimsigh

import "time"

type RegistrationRecord struct {
	DataCenterId string    `bson:"data_center_id"`
	ServiceName  string    `bson:"service_name"`
	InstanceId   string    `bson:"instance_id"`
	IpAddress    string    `bson:"ip_address"`
	Port         int       `bson:"port"`
	CreatedAt    time.Time `bson:"created_at"`
}
