# Sensor Support

This folder contains scripts and tools for sensors and sensor nodes.

## Temperature Sensor to LoRa Pass Through Application

The script `dht2lra.py` is a small Python appliction meant to run on a
Raspberry Pi 3 Model B+. It should be possible to modify to run on any
similar device, although the GPIO pin used to communicate with the
temperature sensor might need to be adjusted. It uses the RPi.GPIO,
PySerial, and simplejson Python modules, all of which are part of the
default Raspberry Pi OS install.

The script can be run from the command line as follows. It will continue
running until killed, e.g. by pressing Ctrl+C.

```
python dht2lra.py
```

When run, the appliction attempts to read temperature and humidity from
a DHT-11 sensor with the DATA line connected to the GPIO14 pin of the
Raspberry Pi. Consult the output of the `pinout` command for the
physical location of the pin. Other pins on the same connector can be
used to provide VCC (5V) and ground to the DHT-11 device.

Each time a reading is successfully retrieved, a line like the one shown
below will be printed, and the data will be sent via LoRa through the LRA-1
dongle attached to one of the Raspberry Pi's USB ports. The script
currently assumes the dongle appears as device `/dev/ttyUSB0`.

```
Sending: {"temp": 20.2, "hum": 42.0}
```

Because the protocol used to communicate with the sensor is somewhat
unreliable, you may see lines like those below. The script will attempt
the reading again after a delay.

```
Not enough data bits: 36
Not enough data bits: 38
Sensor reading failed
```

If the failures persist, stop the script and check the connections of the
sensor.

If the message "Failed to connect to LRA device" appears when the
script starts, check that the LRA-1 is connected to the USB port and that
it is present as `/dev/ttyUSB0` (see below for how to manually connect to
the device).

### Configuring The LoRa Device

The LRA-1 needs to be configured with its own ID and the ID of the LoRa
device connected to the edge node which will receive the sensor data. This
can be done by connecting to the device using `tio` or another similar tool.

The `tio` package is not included in the default Raspberry Pi install, but
can be added with `sudo apt-get install tio`.

If using `tio`, connect to the LRA-1 with the command `sudo tio /dev/ttyUSB0`.
Once connected, you should see a prompt like this:

```
i2-ele LRA1
Ver 1.07.b+
OK
>
```

At the prompt, you can confirm the destination ID for messages `dst`, and
the ID used to identify the sensor node `own`, using the `print` command:

```
>print own
4
OK
>print dst
3
OK
>
```

Set the values appropriately for your configuration, like this:

```
>own=10
OK
>dst=50
OK
>print own
10
OK
>print dst
50
OK
>ssave
OK
>
```

The `ssave` command stores the settings in persistent storage on the LRA-1,
so they do not need to be reconfigured after a power cycle.

### Enabling dht2lra As A Service

On a systemd based system like Ubuntu or Debian, or the Raspberry Pi OS, you
can run the `dht2lra.py` script automatically on startup by copying the
`dht2lra.service` file to `/etc/systemd/system/` and enabling the service:

```
sudo systemctl enable dht2lra.service
```

The service file assumes the `dht2lra.py` file is located in `/home/pi/`. If
the directory name is different, adjust the service file appropriately.

The service can be stopped with `sudo systemctl stop dht2lra.service` and
started manually with `sudo systemctl start dht2lra.service`. It can be
disabled (so it will not automatically start at boot) with
`sudo systemctl disable dht2lra.service`.

