package models

import "time"

type TelemetryEvent struct {
	ID        string         `json:"uniqueId" bson:"_id"`
	User      UserInfo       `json:"user" bson:"user"`
	Platform  PlatformInfo   `json:"platform" bson:"platform"`
	Device    DeviceInfo     `json:"device" bson:"device"`
	EventData map[string]any `json:"eventData" bson:"eventData"`
	TimeStamp time.Time      `json:"timeStamp" bson:"timeStamp"`
	EventType string         `json:"eventType" bson:"eventType"`
}

type UserInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PlatformInfo struct {
	PlatformName  string `json:"platformName" bson:"platformName"`
	Env           string `json:"env" bson:"env"`
	ComponentName string `json:"componentName" bson:"componentName"`
}

type DeviceInfo struct {
	BrowserVersion string `json:"browserVersion" bson:"browserVersion"`
	DeviceType     string `json:"deviceType" bson:"deviceType"`
	Orientation    string `json:"orientation" bson:"orientation"`
	OS             string `json:"os" bson:"os"`
	Browser        string `json:"browser" bson:"browser"`
	Device         string `json:"device" bson:"device"`
	OSVersion      string `json:"osVersion" bson:"osVersion"`
}
