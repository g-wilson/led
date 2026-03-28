# LED Smart Clock

This is my controller code for my Raspberry Pi LED matrix "smart-clock" project.

To run locally, it outputs the LED matrix as a PNG image:

`go run cmd/debug/main.go`

To run on the Pi, it requires compiling the [LED matrix C bindings](https://github.com/hzeller/rpi-rgb-led-matrix), then:

`go run cmd/pi/main.go`

### Config

This project uses dotenv to manage config. Create a `.env` file before running:

```
# Weather
TOMORROWIO_API_KEY=xxxx
WEATHER_LATITUDE=xxxx
WEATHER_LONGITUDE=xxxx
WEATHER_REFRESH=1800

# Home Assistant
HA_URL=http://192.168.1.100:8123
HA_TOKEN=xxxx
HA_SENSORS=sensor.temp,sensor.humidity

# LED hardware (Pi only)
LED_ROWS=32
LED_COLS=64
LED_PWM_BITS=11
LED_PWM_LSB=130
LED_BRIGHTNESS=30
LED_HARDWARE=adafruit-hat

# General
DEBUG=false

# Calendar (optional) — colon-separated list of YAML files to load at runtime
CALENDAR_FILES=/path/to/my-events.yaml:/path/to/work-calendar.yaml
```

### Custom Calendars

Calendar events are defined in YAML files. The repo ships with `internal/calendar/events.yaml` (holidays etc.) and `internal/calendar/f1.yaml` (F1 season), both embedded in the binary.

You can supply additional YAML files at runtime via `CALENDAR_FILES` in `.env`. These files are not committed to the repo, making them suitable for personal or private events.

**YAML format:**

```yaml
events:
  - name: My Event
    time: "2026-06-01T18:00:00Z"
  - name: Birthday
    time: "2026-08-15T09:00:00Z"
    image: /absolute/path/to/icon.png
```

The `image` field is optional and accepts:
- `builtin:f1` or `builtin:xmastree` — built-in icons embedded in the binary
- An absolute file path to a PNG image
- A path relative to the YAML file's directory
