package consulmq

import "testing"

import "time"

func TestRegisterServiceConsul(t *testing.T) {
	mq, err := Connect(Config{
		Address: "172.17.0.2:8500",
	})
	if err != nil {
		t.Error(err)
	}
	err = registerServiceConsul(mq)
	if err != nil {
		t.Error(err)
	}
}

func TestGetIP(t *testing.T) {
	ip, err := getIP("8.8.8.8:53")
	if err != nil {
		t.Error(err)
	}
	if ip == "" {
		t.Errorf("Unable to get local IP!")
	}
	_, err = getIP("failure")
	if err == nil {
		t.Errorf("Something failed!")
	}
}

func TestSetDefaults(t *testing.T) {
	conf := Config{}
	newconf := Config{}
	// test for setting on nul values
	def := setDefaults(&conf, &newconf, defaults)
	if def.Address != defaults["Address"] {
		t.Errorf("Unable to set address default value")
	}
	if def.Datacenter != defaults["Datacenter"] {
		t.Errorf("Unable to set datacenter default value")
	}
	if def.MQName != defaults["MQName"] {
		t.Errorf("Unable to set message queue default value")
	}
	if def.TTL != 87600*time.Hour {
		t.Errorf("Unable to set TTL default value")
	}

	// test for not over-writing given values
	conf = Config{
		Address:    "1.2.3.4:8500",
		Datacenter: "testDatacenter",
		MQName:     "testMQ",
		TTL:        5 * time.Minute,
	}
	def = setDefaults(&conf, &newconf, defaults)
	if def.Address != "1.2.3.4:8500" {
		t.Errorf("Address value was not preserved")
	}
	if def.Datacenter != "testDatacenter" {
		t.Errorf("Datacenter value was not preserved")
	}
	if def.MQName != "testMQ" {
		t.Errorf("MQName value was not preserved")
	}
	if def.TTL != 5*time.Minute {
		t.Errorf("TTLvalue was not preserved")
	}

}
