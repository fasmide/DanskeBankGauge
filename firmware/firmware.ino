#include "FastLED.h"
#include <ESP8266mDNS.h>
#include <WiFiUdp.h>
#include <ArduinoOTA.h>
#include <ArduinoJson.h>
#include <ESP8266HTTPClient.h>
#include <WiFiClient.h>

// secrets includes hostname for the danske bank daemon as well as wifi creds
#include "secrets.h"

#define PIXELPIN D5
#define NUMPIXELS 7
CRGB leds[NUMPIXELS];

// well the gauge is not 100% linear so at 10000 it reads more like 9950 and at 
// 5000 it reads more like 5050 but its good enough for the girls i go out with...
#define FULLSCALE 10400
#define GAUGEPIN 12 // D6

#ifndef STASSID
#define STASSID "some-ssid"
#define STAPSK  "some-password"
#endif

const char* ssid = STASSID;
const char* password = STAPSK;

void setup() {
    Serial.begin(115200);
    Serial.println("Booting");

    // tell FastLED about the LED strip configuration
    FastLED.addLeds<WS2811,PIXELPIN,GRB>(leds, NUMPIXELS).setCorrection(TypicalLEDStrip);
    FastLED.setBrightness(255);
    
    // Do blue when booting
    fill_solid(leds, NUMPIXELS, CRGB::Blue); 
    FastLED.show();

    WiFi.mode(WIFI_STA);
    WiFi.begin(ssid, password);
    while (WiFi.waitForConnectResult() != WL_CONNECTED) {
        Serial.println("Connection Failed! Rebooting...");
        delay(5000);
        ESP.restart();
    }

    // Hostname defaults to esp8266-[ChipID]
    ArduinoOTA.setHostname("DBGauge");

    ArduinoOTA.onStart([]() {
        String type;
        if (ArduinoOTA.getCommand() == U_FLASH) {
            type = "sketch";
        } else { // U_SPIFFS
            type = "filesystem";
        }

        // NOTE: if updating SPIFFS this would be the place to unmount SPIFFS using SPIFFS.end()
        Serial.println("Start updating " + type);
    });
    ArduinoOTA.onEnd([]() {
        Serial.println("\nEnd");
    });
    ArduinoOTA.onProgress([](unsigned int progress, unsigned int total) {
        Serial.printf("Progress: %u%%\r", (progress / (total / 100)));
    });
    ArduinoOTA.onError([](ota_error_t error) {
        Serial.printf("Error[%u]: ", error);
        if (error == OTA_AUTH_ERROR) {
            Serial.println("Auth Failed");
        } else if (error == OTA_BEGIN_ERROR) {
            Serial.println("Begin Failed");
        } else if (error == OTA_CONNECT_ERROR) {
            Serial.println("Connect Failed");
        } else if (error == OTA_RECEIVE_ERROR) {
            Serial.println("Receive Failed");
        } else if (error == OTA_END_ERROR) {
            Serial.println("End Failed");
        }
    });
    ArduinoOTA.begin();
    Serial.println("Ready");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
  
}

const unsigned long fiveMinutes = 5 * 60 * 1000UL;

// when testing - update every second
// const unsigned long fiveMinutes = 1 * 1000UL;

// initialize such that a reading is due the first time through loop()
static unsigned long lastSampleTime = 0 - fiveMinutes;  

int balance;
int allowance;

void loop() {
    ArduinoOTA.handle();

    unsigned long now = millis();
    if (now - lastSampleTime >= fiveMinutes)
    {
        lastSampleTime += fiveMinutes;
        Serial.println("Time for updating!");
        bool success = update();

        if (!success) {
            analogWrite(GAUGEPIN, 0);
            fill_solid(leds, NUMPIXELS, CRGB::Blue);
            FastLED.show();
            return;
        }

        if (allowance >= 0) {
            fill_solid(leds, NUMPIXELS, CRGB::Green); 
        } else {
            fill_solid(leds, NUMPIXELS, CRGB::Red); 
        }
        // we are using software PWM so we should not interrupt FastLED while its updating
        // its so fast its not noticable anyways
        analogWrite(GAUGEPIN, 0);
        FastLED.show();  
        analogWrite(GAUGEPIN, scale(balance));
    }
    

}
bool update() {    
    
    WiFiClient client;
    HTTPClient http;

    bool success = false;

    Serial.print("[HTTP] begin...\n");
    if (http.begin(client, BALANCEURL)) {  // HTTP


        Serial.print("[HTTP] GET...\n");
        // start connection and send HTTP header
        int httpCode = http.GET();

        // httpCode will be negative on error
        if (httpCode > 0) {
            // HTTP header has been send and Server response header has been handled
            Serial.printf("[HTTP] GET... code: %d\n", httpCode);

            // file found at server
            if (httpCode == HTTP_CODE_OK || httpCode == HTTP_CODE_MOVED_PERMANENTLY) {
                const size_t capacity = JSON_OBJECT_SIZE(2) + 20;
                DynamicJsonDocument doc(capacity);

                deserializeJson(doc, http.getString());

                balance = doc["balance"]; // 10500
                allowance = doc["allowance"]; // 4427
                success = true;
            }
        } else {
                Serial.printf("[HTTP] GET... failed, error: %s\n", http.errorToString(httpCode).c_str());
        }

        http.end();
    } else {
        Serial.printf("[HTTP} Unable to connect\n");
    }
    return success;
}
int scale(int input) {
    if (input <= 0) {
        return 0;
    }
    if (input >= FULLSCALE) {
        return FULLSCALE;
    }

    // map(value, fromLow, fromHigh, toLow, toHigh)
    return map(input, 0, FULLSCALE, 0, 1024);
}