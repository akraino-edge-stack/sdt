// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Fujitsu Limited
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"time"
	"io"
	"fmt"
	"strconv"
	"strings"
	"github.com/tarm/serial"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	ModemConfigName         = "modem"
	TxPowerConfigName       = "pwr"
	SpreadFactorConfigName  = "sf"
	BandwidthConfigName     = "bw"
	ChannelConfigName       = "ch"
	FrequencyConfigName     = "frq"
	GroupIDConfigName       = "gid"
	OwnIDConfigName         = "own"
	DestinationIDConfigName = "dst"
	ControlFlagsConfigName  = "ctrl"
)

type LoRaMessage struct {
	id int
	msg string
}

type LoRaConfig struct {
	modem         int
	txPower       int
	spreadFactor  int
	bandwidth     int
	channel       int
	frequency     int
	groupID       int
	ownID         int
	destinationID int
	controlFlags  int
}

type LoRaTransport struct {
	lc           logger.LoggingClient
	txch         chan LoRaMessage
	rxch         chan LoRaMessage
	port         *serial.Port
	serialConfig *serial.Config
	config       LoRaConfig
}

func (self *LoRaTransport) ReadLines() ([]string, error) {
        var buf []byte
        for {
                b := make([]byte, 512)
                n, err := self.port.Read(b)
		if err != nil {
			// EOF returned on timeout
			if err != io.EOF || n != 0 {
				return nil, err
			}
		}
		if n == 0 {
			break
		}
		buf = append(buf, b[:n]...)
        }
	if len(buf) > 0 {
		lines := strings.Split(strings.ReplaceAll(string(buf),"\r","\n"),"\n")
		self.lc.Debugf("ReadLines: %v", lines)
		return lines, nil
	}
	return make([]string, 0), nil
}

// Reads any waiting messages and forwards them to the channel
func (self *LoRaTransport) ReadData() error {
	lines, err := self.ReadLines()
	if err == nil && len(lines) > 0 {
		for _, l := range lines {
			if len(l) > 0 && strings.HasPrefix(l, "@") {
				// Looks like a received message
				self.lc.Debugf("Transport RX: %s", l)
				parts := strings.SplitN(l, ",", 3)
				if len(parts) == 3 {
					var msg LoRaMessage
					if msg.id, err = strconv.Atoi(parts[1]); err == nil {
						msg.msg = strings.TrimSpace(parts[2])
						self.rxch <- msg
					} else {
						// Ignore conversion errors
						err = nil
					}
				}
			}
		}
	}
        return err
}

func (self *LoRaTransport) WriteData(s string) error {
	self.lc.Debugf("Transport TX: %s", s)
	_, err := self.port.Write([]byte(s+"\r"))
	if err != nil {
		self.port.Close()
		self.port = nil
		return err
	}
	return nil
}

func (self *LoRaTransport) SendBreak() {
	self.lc.Debugf("BREAK")
	// Unfortunately the only way to break out of "comm" is disconnect
	// (or use the break signal, but serial does not support that)
	// _, err := self.port.Write([]byte{0x03})
	// if err != nil {
	//	self.lc.Errorf("Error sending break: %s", err.Error())
	// }

	// FIXME: This is messy and unsatisfying
	self.port.Close()
	var err error
        self.port, err = serial.OpenPort(self.serialConfig)
	if err != nil {
		self.lc.Errorf("Failed to reopen serial port")
		return
	}
	// Read until we see the ">" prompt
	for i := 0; i < 15; i = i + 1 {
                b := make([]byte, 16)
		n, rderr := self.port.Read(b)
		if rderr != nil {
			// EOF returned on timeout
			if rderr != io.EOF || n != 0 {
				self.lc.Errorf("Error waiting for prompt")
				break
			}
		}
		if n > 0 && b[n-1] == byte('>') {
			break
		}
	}
}

func (self *LoRaTransport) Close() {
	self.lc.Info("Closing LoRa transport")
	if self.port != nil {
		self.SendBreak()
		self.port.Close()
		self.port = nil
	}
	close(self.rxch)
}

func (self *LoRaTransport) GetIntVariable(name string) (int, error) {
	err := self.WriteData("print " + name)
	var n int
	if err == nil {
		lines, rderr := self.ReadLines()
		if rderr == nil {
			if len(lines) > 0 {
				for i, line := range lines {
					// Skip first line (echo of command)
					// and blank lines, including
					// extras from CR-LF
					if i > 0 && len(line) > 0 && line[0] != '\n' {
						n, err = strconv.Atoi(line)
						break
					}
				}
			} else {
				err = fmt.Errorf("Read failed")
			}
		} else {
			err = rderr
		}
	}
	return n, err
}

func (self *LoRaTransport) SetIntVariable(name string, value int) error {
	err := self.WriteData(name + "=" + strconv.Itoa(value))
	if err != nil {
		return err
	}
	_, rderr := self.ReadLines()
	if rderr != nil {
		return rderr
	}
	newvalue, geterr := self.GetIntVariable(name)
	if geterr != nil {
		return geterr
	}
	if newvalue != value {
		return fmt.Errorf("Value of %s not set to %d (%d instead)", name, value, newvalue)
	}
	return nil
}

func (self *LoRaTransport) DumpSettings(settingsMap map[string]int) {
	for k, v := range settingsMap {
		self.lc.Infof("%8s : %d", k, v)
	}
}

func (self *LoRaTransport) GetAndUpdateSettings() {
	settingsMap := make(map[string]int)
	settingsMap[ModemConfigName]         = self.config.modem
	settingsMap[TxPowerConfigName]       = self.config.txPower
	settingsMap[SpreadFactorConfigName]  = self.config.spreadFactor
	settingsMap[BandwidthConfigName]     = self.config.bandwidth
	settingsMap[ChannelConfigName]       = self.config.channel
	settingsMap[FrequencyConfigName]     = self.config.frequency
	settingsMap[GroupIDConfigName]       = self.config.groupID
	settingsMap[OwnIDConfigName]         = self.config.ownID
	settingsMap[DestinationIDConfigName] = self.config.destinationID
	settingsMap[ControlFlagsConfigName]  = self.config.controlFlags

	for k, _ := range settingsMap {
		v, err := self.GetIntVariable(k)
		if err != nil {
			self.lc.Warnf("Failed to read %s", k)
		} else {
			settingsMap[k] = v
		}
	}
	self.DumpSettings(settingsMap)

	modified := false
	self.config.modem = settingsMap[ModemConfigName]
	self.config.txPower = settingsMap[TxPowerConfigName]
	self.config.spreadFactor = settingsMap[SpreadFactorConfigName]
	self.config.bandwidth = settingsMap[BandwidthConfigName]
	if self.config.channel != settingsMap[ChannelConfigName] {
		modified = true
		self.lc.Infof("Set device channel to %d to match config", self.config.channel)
		err := self.SetIntVariable(ChannelConfigName, self.config.channel)
		if err != nil {
			self.lc.Errorf("Failed to update channel: %s", err.Error())
		}

	}
	self.config.frequency = settingsMap[FrequencyConfigName]
	if self.config.groupID != settingsMap[GroupIDConfigName] {
		modified = true
		self.lc.Infof("Set group ID to %d to match config", self.config.groupID)
		err := self.SetIntVariable(GroupIDConfigName, self.config.groupID)
		if err != nil {
			self.lc.Errorf("Failed to update group ID: %s", err.Error())
		}
	}
	if self.config.ownID != settingsMap[OwnIDConfigName] {
		modified = true
		self.lc.Infof("Set station ID to %d to match config", self.config.ownID)
		err := self.SetIntVariable(OwnIDConfigName, self.config.ownID)
		if err != nil {
			self.lc.Errorf("Failed to update station ID: %s", err.Error())
		}
	}
	self.config.destinationID = settingsMap[DestinationIDConfigName]
	self.config.controlFlags = settingsMap[ControlFlagsConfigName]
	if modified {
		err := self.WriteData("ssave")
		if err != nil {
			self.lc.Errorf("Failed to save updated config: %s", err.Error())
		}
	}
	self.lc.Infof("This station ID: %d", self.config.ownID)
}

func (self *LoRaTransport) Initialize() error {
        self.serialConfig = &serial.Config{
                Name:        "/dev/ttyUSB0",
                Baud:        115200,
                ReadTimeout: time.Second * 1,
        }
        retrys := 0
	var err error
        for self.port, err = serial.OpenPort(self.serialConfig); err != nil; self.port, err = serial.OpenPort(self.serialConfig) {
                if retrys % 100 == 0 {
			self.lc.Infof("Can't open serial port: %s", err.Error())
                        retrys = 0
                }
                // Check if the channel has been closed
                // And also drain any messages clogging it up
                select {
                case _, ok := <-self.txch:
                        if !ok {
                                // Stop if the driver stopped
				self.port = nil
                                return nil
                        }
                default:
                }
                time.Sleep(time.Second * 5)
                retrys = retrys + 1
        }

	// Read until we see the ">" prompt
        for {
                b := make([]byte, 16)
		n, rderr := self.port.Read(b)
		if rderr != nil {
			// EOF returned on timeout
			if rderr != io.EOF || n != 0 {
				self.Close()
				return rderr
			}
		}
		if n > 0 && b[n-1] == byte('>') {
			break
		}
                // Again check if the channel has been closed
                // and drain any messages clogging it up
                select {
                case _, ok := <-self.txch:
                        if !ok {
                                // Stop if the driver stopped
				self.Close()
                                return nil
                        }
                default:
                }
        }
	self.GetAndUpdateSettings()
	wrerr := self.WriteData("comm")
	if wrerr != nil {
		return wrerr
	}

	return nil
}

func (self *LoRaTransport) Reset() error {
	self.lc.Info("Resetting transport")
	self.SendBreak()
	wrerr := self.WriteData("comm")
	if wrerr != nil {
		return wrerr
	}
	return nil
}

func (self *LoRaTransport) SetDestination(id int) error {
	if id != self.config.destinationID {
		// Break communication
		self.SendBreak()
		_, err := self.ReadLines()
		if err != nil {
			return err
		}
		err = self.SetIntVariable(DestinationIDConfigName, id)
		if err != nil {
			return err
		} else {
			self.config.destinationID = id
		}
		// Restart communication
		err = self.WriteData("comm")
		if err != nil {
			return err
		}
		_, err = self.ReadLines()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *LoRaTransport) Run() error {
	defer self.Close()
	err := self.Initialize()
	if err == nil && self.port == nil {
		err = fmt.Errorf("Transport not ready")
	}
	if err != nil {
		self.lc.Errorf("Failed to initialize transport: %s", err.Error())
		return err
	}
	self.lc.Info("LoRa transport initialized")
Forever:
	for {
		select {
		case msg, ok := <-self.txch:
			if !ok {
				// Driver closed the channel
				break Forever
			}
			// Make sure the transport is idle/ready
			err = self.ReadData()
			if err != nil {
				break Forever
			}
			// Message from the driver to write to LoRa
			err = self.SetDestination(msg.id)
			if err == nil {
				err = self.WriteData(msg.msg)
				if err != nil {
					break Forever
				}
			} else {
				self.lc.Errorf("Error setting destination: %s", err.Error())
				err = self.Reset()
				if err != nil {
					break Forever
				}
			}
		default:
			// Check for messages to read
			err = self.ReadData()
			if err != nil {
				break Forever
			}
		}
	}
	if err != nil {
		self.lc.Errorf("Fatal transport error: %s", err.Error())
	}
	self.lc.Info("LoRa transport stopped")
	return err
}

func (self *LoRaTransport) Start() error {
	go self.Run()
	return nil
}

func NewLoRaTransport(lc logger.LoggingClient, rxch chan LoRaMessage, txch chan LoRaMessage) *LoRaTransport {
	t := new(LoRaTransport)
	t.lc = lc
	t.txch = txch
	t.rxch = rxch
	// Default config values
	// NOTE: Most of these are read from the device
	// Only ownID, groupID and channel are written to the device
	t.config.modem         = 1          // 1: LoRa, 0: FSK
	t.config.txPower       = 13         // -4 to 13 dBm
	t.config.spreadFactor  = 10         // 7 to 12
	// Bandwidth 6: 62.5, 7: 125, 8: 250, 9: 500 kHz
	t.config.bandwidth     = 7
	t.config.channel       = 36         // 24 to 61
	t.config.frequency     = 923000000  // hZ
	t.config.groupID       = 0
	t.config.ownID         = 0
	t.config.destinationID = 1
	t.config.controlFlags  = 0
	return t
}
