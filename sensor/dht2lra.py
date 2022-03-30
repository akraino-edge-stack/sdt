#!/usr/bin/python3
# -*- coding: utf-8 -*-

# Periodically read temperature and humidity data from DHT-11 sensor
# convert to JSON format and transmit via LRA1 LoRa device on serial
# port (default /dev/ttyUSB0).
#
# The JSON format is two values "temp" (temperature in degrees C) and
# "hum" (relative humidity percent), both of which are floating point
# numbers, e.g.:
#  { "temp": 20.0, "hum": 44.0 }
#
# COPYRIGHT 2022 FUJITSU LIMITED
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import RPi.GPIO as gpio
import simplejson as json
import serial
import io
import time
import random
import numbers

MAX_READ=100

class Dht:
    def __init__(self):
        # Constructor
        # The DHT-11 data pin is connected to J8 pin 14
        self.pin = 14
        self.data = []
        gpio.setmode(gpio.BCM)

    # From DHT-11 datasheet
    #
    #  Data consists of decimal and integral parts. A complete data
    #  transmission is 40bit, and the sensor sends higher data bit first.
    #  Data format:
    #    8bit integral RH data + 8bit decimal RH data +
    #    8bit integral T data + 8bit decimal T data +
    #    8bit check sum.
    #  If the data transmission is right, the check-sum should be the last
    #  8bit of "8bit integral RH data + 8bit decimal RH data +
    #  8bit integral T data + 8bit decimal T data".
    def check(self):
        # The data parameter is a list of five integers, representing
        # the five bytes received from the sensor.
        if len(self.data) != 5:
            return False

        # Check that the checksum in the last byte is correct.
        return (((self.data[0] + self.data[1] + self.data[2] + self.data[3]) & 255) == self.data[4])

    def humidity(self):
        return self.data[0] + (self.data[1] / 10.0)

    def temperature(self):
        return self.data[2] + (self.data[3] / 10.0)


    # From DHT-11 datasheet
    #
    #  When the communication between MCU and DHT11 begins, the programme
    #  of MCU will set Data Single-bus voltage level from high to low
    #  and this process must take at least 18ms to ensure DHT’s detection
    #  of MCU's signal, then MCU will pull up voltage and wait 20-40us for
    #  DHT’s response.
    def wake(self):
        gpio.setup(self.pin, gpio.OUT)
        gpio.output(self.pin, gpio.HIGH)
        time.sleep(0.05)
        gpio.output(self.pin, gpio.LOW)
        time.sleep(0.02)
        gpio.setup(self.pin, gpio.IN, pull_up_down=gpio.PUD_UP)

    def read_signal(self):
        # Reads a string of signal values until they stop changing.
        u = 0
        last = -1
        values = []
        start_time = time.perf_counter()
        while True:
            v = gpio.input(self.pin)
            values.append(v)
            if v != last:
                last = v
                u = 0
            else:
                u = u + 1
                if u > MAX_READ:
                    break
        end_time = time.perf_counter()
        self.bit_time = (end_time - start_time) / len(values)
        return values

    def read(self):
        # Returns True on success, False on failure
        u = 0
        bits = []
        self.wake()
        values = self.read_signal()
        last = None
        started_data = False
        for v in values:
            if started_data:
                if v != last:
                    last = v
                    if v == gpio.HIGH:
                        # Start of bit data high
                        u = 0
                    else:
                        # End of bit
                        high_time = u * self.bit_time // 1.0e-6
                        if high_time > 40:
                            bits.append(1)
                        else:
                            bits.append(0)
                        if len(bits) == 40:
                            break
                else:
                    u = u + 1
            else:
                if last == gpio.LOW:
                    # We've seen the initial low signal, waiting for high.
                    if v == gpio.HIGH:
                        last = v
                elif last == gpio.HIGH:
                    # We've seen the initial high signal, wait for data's
                    # first low
                    if v == gpio.LOW:
                        started_data = True
                        last = v
                else: 
                    # Watch for initial low signal
                    if v == gpio.LOW:
                        last = v

        if len(bits) != 40:
            print("Not enough data bits: " + str(len(bits)))
            return False

        # Clear old data
        self.data = []
        for i in range(0, 5):
            self.data.append(0)

        # Assemble bytes
        byte = 0
        for i in range(0, len(bits)):
            if bits[i]:
                self.data[byte] = self.data[byte] | (1 << (7 - (i % 8)))
            if ((i+1) % 8) == 0:
                byte = byte + 1
        return self.check()

    def close(self):
        gpio.cleanup()

class Lra:
    def __init__(self):
        # TODO: Support other serial ports
        self.ser = serial.Serial("/dev/ttyUSB0",115200, timeout=1)
        self.sio = io.TextIOWrapper(io.BufferedRWPair(self.ser, self.ser))
        tries = 0
        while True:
            line = self.sio.readline()
            if line == ">":
                break
            tries += 1
            if tries % 10 == 0:
                self.ser.send_break()
            if tries == 29:
                print("Failed to connect to LRA device")
                self.ser.close()
                return
        # Enter communication mode
        self.sio.write("comm\r")
        self.sio.flush()

    def read(self):
        # Read any received messages
        msgs = []
        while True:
            line = self.sio.readline()
            if line.endswith("\n") and line.startswith("@"):
                parts = line.split(',',2)
                if len(parts) == 3:
                    msgs.append(parts[2].rstrip('\r\n'))
            elif line == "":
                break
        return msgs

    def send(self, msg):
        print("Sending: " + msg)
        self.sio.write(msg + "\r")
        self.sio.flush()

    def close(self):
        # Send BREAK to end "comm" command
        self.ser.send_break()
        while True:
            line = self.sio.readline()
            if line == "" or line == ">":
                break
        self.ser.close()

def get_base_interval(interval):
    return interval * 0.9

def get_jitter(interval, base):
    return (interval - base) * 2

def main():
    dht = Dht()
    lra = Lra()
    interval = 5
    base_interval = get_base_interval(interval)
    jitter = get_jitter(interval, base_interval)
    retry_sleep = 1
    min_interval = 2
    max_interval = 600
    random.seed()

    try:
        while True:
            # Check for incoming messages
            msgs = lra.read()
            if len(msgs) > 0:
                for msg in msgs:
                    print("Received: " + msg)
                    try:
                        o = json.loads(msg)
                        if 'int' in o:
                            # Set interval
                            ni = o['int']
                            if isinstance(ni, numbers.Number) and ni >= min_interval and ni <= max_interval:
                                print("New interval: " + str(ni))
                                interval = ni
                                base_interval = get_base_interval(interval)
                                jitter = get_jitter(interval, base_interval)
                    except json.JSONDecodeError:
                        print("Message could not be parsed")
            next_interval = interval + jitter * random.random()
            skip = False
            while not dht.read():
                if next_interval < min_interval:
                    skip = True
                    break
                next_retry_sleep = retry_sleep + jitter * random.random()
                next_interval = next_interval - next_retry_sleep
                time.sleep(next_retry_sleep)
            if skip != True:
                msg = json.dumps({"temp": dht.temperature(), "hum": dht.humidity()})
                lra.send(msg)
            else:
                print("Sensor reading failed")
            if next_interval > 0.0:
                time.sleep(next_interval)
    except KeyboardInterrupt:
        print('Stop')

    dht.close()
    lra.close()

if __name__ == '__main__':
    main()

