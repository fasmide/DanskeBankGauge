#include "FastLED.h"


#define PIXELPIN D5
#define NUMPIXELS 7

// well the gauge is not 100% linear so at 10000 it reads more like 9950 and at 
// 5000 it reads more like 5050 but its good enough for the girls i go out with...
#define FULLSCALE 10400
#define GAUGEPIN 12 // D6
 
CRGB leds[NUMPIXELS];

int outputValue = 0;

void setup() {

    // tell FastLED about the LED strip configuration
    FastLED.addLeds<WS2811,PIXELPIN,GRB>(leds, NUMPIXELS).setCorrection(TypicalLEDStrip);
    FastLED.setBrightness(255);

    // initialize serial communications at 9600 bps:
    Serial.begin(9600);
}

void loop() {
    if (Serial.available()) {

        outputValue = Serial.parseInt();
    }
    analogWrite(GAUGEPIN, scale(outputValue));    
    
    EVERY_N_MILLISECONDS( 16 ) { animate(); FastLED.show(); }  

}
int counter = 0;
void animate() {
    // a colored dot sweeping back and forth, with fading trails
    fadeToBlackBy( leds, NUMPIXELS, 10);
    counter ++;
    if (counter > NUMPIXELS) {
        counter = 0;
    } 
    leds[counter] += CHSV( 0, 255, 255);

}

// we want to drift between 120 degrees and 0 on the hue weel (green to red)
int scaleColor(int input) {
    if (input <= -1000) {
        return 0;
    }
    if (input >= 1000) {
        return 120;
    }
    return map(input, -1000, 1000, 0, 120);
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