#include <Adafruit_NeoPixel.h>


#define PIN D5
// How many NeoPixels are attached to the Arduino?
#define NUMPIXELS 7
 
Adafruit_NeoPixel pixels = Adafruit_NeoPixel(7, PIN, NEO_GRB + NEO_KHZ800);

const int analogOutPin = 12; // D6

int outputValue = 0;
uint32_t color = 0;

void setup() {
    // initialize serial communications at 9600 bps:
    Serial.begin(9600);
    pixels.begin();
}

void loop() {
    outputValue = random(0, 1024);
    analogWrite(analogOutPin, outputValue);

    Serial.print("output = ");
    Serial.println(outputValue);

    color = pixels.Color(random(0, 255), random(0, 255),random(0, 255));
    for(int i=0;i<NUMPIXELS;i++)
    {
        pixels.setPixelColor(i, color);
    }
    pixels.show();
    delay(500);
}
