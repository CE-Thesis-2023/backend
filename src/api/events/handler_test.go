package eventsapi_test

import (
	"encoding/json"
	"fmt"
	"testing"

	eventsapi "github.com/CE-Thesis-2023/backend/src/api/events"
)

func serializeTopic(topic *eventsapi.Topic) string {
	s, _ := json.Marshal(topic)
	return string(s)
}

func TestParseTopic(t *testing.T) {
	t.Run("OpenGate events", func(t *testing.T) {
		exampleTopic := "opengate/test-device-01/events"

		topic, err := eventsapi.ToTopic(exampleTopic)
		if err != nil {
			t.Fatalf("ToTopic(%q) returned error: %v", exampleTopic, err)
		}

		if topic.Type != "opengate" {
			t.Errorf("Expected topic.Type to be %q, got %q", "opengate", topic.Type)
		}

		fmt.Println(serializeTopic(topic))
	})

	t.Run("OpenGate snapshot", func(t *testing.T) {
		exampleTopic := "opengate/test-device-01/ip_camera_031713080742.601523-nrzoin/snapshot"

		topic, err := eventsapi.ToTopic(exampleTopic)
		if err != nil {
			t.Fatalf("ToTopic(%q) returned error: %v", exampleTopic, err)
		}

		if topic.Type != "opengate" {
			t.Errorf("Expected topic.Type to be %q, got %q", "opengate", topic.Type)
		}

		fmt.Println(serializeTopic(topic))
	})

	t.Run("OpenGate available", func(t *testing.T) {
		exampleTopic := "opengate/test-device-01/available"

		topic, err := eventsapi.ToTopic(exampleTopic)
		if err != nil {
			t.Fatalf("ToTopic(%q) returned error: %v", exampleTopic, err)
		}

		if topic.Type != "opengate" {
			t.Errorf("Expected topic.Type to be %q, got %q", "opengate", topic.Type)
		}

		fmt.Println(serializeTopic(topic))
	})

	t.Run("OpenGate states", func(t *testing.T) {
		exampleTopics := []string{
			"opengate/test-device-01/ip_camera_03/audio/state",
			"opengate/test-device-01/ip_camera_03/snapshots/state",
			"opengate/test-device-01/ip_camera_03/recordings/state",
			"opengate/test-device-01/ip_camera_03/detect/state",
			"opengate/test-device-01/ip_camera_03/motion/state",
			"opengate/test-device-01/ip_camera_03/ptz_autotracker/state",
			"opengate/test-device-01/ip_camera_03/improve_contrast/state",
			"opengate/test-device-01/ip_camera_03/birdseye/state",
			"opengate/test-device-01/ip_camera_03/motion_contour_area/state",
			"opengate/test-device-01/ip_camera_03/motion_threshold/state",
			"opengate/test-device-01/ip_camera_03/birdseye_mode/state",
		}
		for _, exampleTopic := range exampleTopics {
			topic, err := eventsapi.ToTopic(exampleTopic)
			if err != nil {
				t.Fatalf("ToTopic(%q) returned error: %v", exampleTopic, err)
			}

			if topic.Type != "opengate" {
				t.Errorf("Expected topic.Type to be %q, got %q", "opengate", topic.Type)
			}

			fmt.Println(serializeTopic(topic))
		}
	})
}
