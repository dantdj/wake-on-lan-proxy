package esxi

import (
	"github.com/dantdj/wake-on-lan-proxy/internal/wakeonlan"
	"github.com/rs/zerolog/log"
	"github.com/sfreiberg/simplessh"
)

type Connection struct {
	URL        string
	MACAddress string
	Username   string
	Password   string
	PoweredOn  bool
}

func New(url, username, password, mac string) Connection {
	return Connection{
		URL:        url,
		Username:   username,
		Password:   password,
		MACAddress: mac,
	}
}

func (ec *Connection) TurnOnServer() error {
	return wakeonlan.SendWolPacket(ec.MACAddress)
}

func (ec *Connection) TurnOffServer() error {
	return ec.sendSSHCommand("esxcli system shutdown poweroff --reason 'routine shutdown'")
}

// Send a generic SSH command to the current ESXi server
func (ec *Connection) sendSSHCommand(command string) error {
	client, err := simplessh.ConnectWithPassword(ec.URL, ec.Username, ec.Password)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Exec(command)
	if err != nil {
		return err
	}

	return nil
}

func (ec *Connection) ServerReachable() bool {
	err := ec.sendSSHCommand("esxcli --version")
	if err != nil {
		log.Error().
			Err(err).
			Str("username", ec.Username).
			Str("password", ec.Password).
			Str("host", ec.URL).
			Msg("error when sending SSH message")
		return false
	}

	return true
}
