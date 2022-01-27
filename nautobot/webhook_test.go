package nautobot

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

// CreateByteIPaddressArray Func convert json to []byte
func CreateByteIPaddressArray(ip_add *IPaddress) []byte {
	byteArray, err := json.Marshal(ip_add)
	if err != nil {
		log.Fatal(err)
	}

	return byteArray
}

// TestNewIPaddress func to test NewIPaddress
func TestNewIPaddress(t *testing.T) {
	// Create Testing Data
	ip_add := &IPaddress{
		Event: "created",
	}
	ip_add.Data.Family.Value = 4
	ip_add.Data.Address = "192.168.1.1/24"
	ip_add.Data.Status.Value = "active"
	ip_add.Data.Dns_name = "test.if.lastmile.sk."

	// Marshal Testing data,
	// Test Unmarshal NewIPaddress function
	exp := NewIPaddress(CreateByteIPaddressArray(ip_add))

	// Compare testing structure
	if reflect.DeepEqual(ip_add, exp) {
		// Test Pass!
		t.Log("Unmarshal IPaddress struct. Get: ", exp)
	} else {
		// Test Fail!
		t.Fatal("Unable unmarshal IPAddress struct. Get: ", exp)
	}
}
