# LED Smart Clock

This is my controller code for my Raspberry Pi LED matrix "smart-clock" project.

To run locally, it outputs the LED matrix as a PNG image:

`go run cmd/debug/main.go`

To run on the Pi, it requires compiling the [LED matrix C bindings](https://github.com/hzeller/rpi-rgb-led-matrix), then:

`go run cmd/pi/main.go`

### Config

This project uses dotenv to manage config. Create a `.env` file before running:

```
# Weather settings
DARKSKY_API_KEY=xxxx
WEATHER_LATITUDE=xxxx
WEATHER_LONGITUDE=xxxx
WEATHER_REFRESH=1800

# LED settings
LED_REFRESH=1000
LED_ROWS=32
LED_COLS=64
LED_PWM_BITS=11
LED_PWM_LSB=130
LED_BRIGHTNESS=30
LED_HARDWARE=adafruit-hat
```
