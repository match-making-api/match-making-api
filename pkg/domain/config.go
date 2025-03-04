package domain

// SteamConfig contains the necessary configuration for connecting to the Steam authentication service.
type SteamConfig struct {
	SteamKey    string
	PublicKey   string
	Certificate string
	VHashSource string
}

// AuthConfig contains the necessary configuration for connecting to the authentication service.
type AuthConfig struct {
	SteamConfig SteamConfig
}

// MongoDBConfig contains the necessary configuration for connecting to the MongoDB database.
type MongoDBConfig struct {
	DBName      string
	URI         string
	PublicKey   string
	Certificate string
}

// Config contains the entire configuration for the application.
type Config struct {
	Auth    AuthConfig
	MongoDB MongoDBConfig
	Kafka   KafkaConfig
}

// KafkaConfig contains the necessary configuration for connecting to the Kafka service.
type KafkaConfig struct {
	// Kafka bootstrap brokers to connect to, as a comma separated list (ie: "kafka1:9092,kafka2:9092")
	Brokers string

	// Kafka cluster version (ie.: "2.1.1", "2.2.2", "2.3.0", ...)
	Version string

	// Kafka consumer group definition (ie: consumer group name)
	Group string

	// Kafka topics to be consumed, as a comma separated list (ie: "topic1,topic2,topic3")
	Topics string

	// Consumer group partition assignment strategy (ie: range, roundrobin, sticky)
	AssignmentStrategy string

	// Kafka consumer consume initial offset from oldest (default: true)
	Oldest bool

	// Sarama logging (default: false)
	Verbose bool
}
