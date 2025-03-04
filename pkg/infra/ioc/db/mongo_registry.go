package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gofrs/uuid"
	"github.com/golobby/container/v3"
	common "github.com/leet-gaming/match-making-api/pkg/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	tUUID       = reflect.TypeOf(uuid.UUID{})
	uuidSubtype = byte(0x04)

	MongoRegistry = bson.NewRegistry()
)

func init() {
	MongoRegistry.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	MongoRegistry.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))
}

// uuidEncodeValue encodes a UUID value into BSON format.
//
// Parameters:
//   - ec: The EncodeContext, which provides contextual information for the encoding process.
//   - vw: The ValueWriter interface used to write the encoded value.
//   - val: The reflect.Value containing the UUID to be encoded.
//
// Returns:
//
//	An error if the encoding process fails, or nil if successful.
func uuidEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{Name: "uuidEncodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}
	b := val.Interface().(uuid.UUID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

// uuidDecodeValue decodes a BSON value into a UUID.
//
// Parameters:
//   - dc: The DecodeContext, which provides contextual information for the decoding process.
//   - vr: The ValueReader interface used to read the BSON value.
//   - val: The reflect.Value where the decoded UUID will be stored.
//
// Returns:
//   - An error if the decoding process fails, or nil if successful.
func uuidDecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{Name: "uuidDecodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		err = vr.ReadNull()
	case bson.TypeUndefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uuid2))
	return nil
}

// InjectMongoDB registers a MongoDB client as a singleton in the provided container.
//
// Parameters:
//   - c: A container.Container instance where the MongoDB client will be registered.
//
// Returns:
//   - error: An error if the MongoDB client registration or connection fails, nil otherwise.
func InjectMongoDB(c container.Container) error {
	err := c.Singleton(func() (*mongo.Client, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for mongo.Client.", "err", err)
			return nil, err
		}

		mongoOptions := options.Client().ApplyURI(config.MongoDB.URI).SetRegistry(MongoRegistry).SetMaxPoolSize(100)

		client, err := mongo.Connect(context.TODO(), mongoOptions)

		if err != nil {
			slog.Error("Failed to connect to MongoDB.", "err", err)
			return nil, err
		}

		return client, nil
	})

	if err != nil {
		slog.Error("Failed to load mongo.Client.")
		return err
	}

	return nil
}
