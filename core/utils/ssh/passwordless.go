package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

func loadRemoteSigners(home string) ([]gossh.Signer, error) {
	var signers []gossh.Signer
	var loadErr error
	for _, name := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
		data, err := os.ReadFile(filepath.Join(home, ".ssh", name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("load %s: %w", name, err))
			continue
		}
		signer, err := gossh.ParsePrivateKey(data)
		if err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("parse %s: %w", name, err))
			continue
		}
		signers = append(signers, signer)
	}
	if len(signers) == 0 {
		return nil, errors.Join(fmt.Errorf("no usable default SSH private keys found"), loadErr)
	}
	return signers, nil
}

func remoteKeyPath(home string) (string, error) {
	dir := filepath.Join(home, ".ssh")
	var loadErr error
	for _, name := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("load %s: %w", name, err))
			continue
		}
		if _, err := gossh.ParsePrivateKey(data); err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("parse %s: %w", name, err))
			continue
		}
		return path, nil
	}
	if loadErr != nil {
		return "", loadErr
	}
	return filepath.Join(dir, "id_ed25519"), nil
}

func LoadRemoteKey(create bool) ([]byte, error) {
	current, err := user.Current()
	if err != nil {
		return nil, err
	}
	name, err := remoteKeyPath(current.HomeDir)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(name)
	if !create {
		return os.ReadFile(name)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(name+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := lock.Close(); err != nil {
			global.LOG.Warnf("close SSH key lock file %s failed: %v", lock.Name(), err)
		}
	}()
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil {
		return nil, err
	}
	defer func() {
		if err := unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
			global.LOG.Warnf("unlock SSH key lock file %s failed: %v", lock.Name(), err)
		}
	}()
	data, err := os.ReadFile(name)
	if errors.Is(err, os.ErrNotExist) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if output, err := exec.CommandContext(ctx, "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", name).CombinedOutput(); err != nil {
			return nil, fmt.Errorf("ssh-keygen failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
		data, err = os.ReadFile(name)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	signer, err := gossh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(name, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(name+".pub", gossh.MarshalAuthorizedKey(signer.PublicKey()), 0644); err != nil {
		return nil, err
	}
	return data, nil
}

func ConfigurePasswordless(info ConnInfo) error {
	probe := info
	probe.AuthMode = "passwordless"
	client, err := NewClient(probe)
	if err == nil {
		client.Close()
		return nil
	}
	key, err := LoadRemoteKey(true)
	if err != nil {
		return err
	}
	if info.UseProxy {
		client, err := NewClient(info)
		if err != nil {
			return err
		}
		defer client.Close()
		signer, err := gossh.ParsePrivateKey(key)
		if err != nil {
			return err
		}
		if err := client.installRemoteAuthorization(signer.PublicKey()); err != nil {
			return err
		}
		verified, err := NewClient(probe)
		if err != nil {
			return fmt.Errorf("public key installed but passwordless verification failed: %w", err)
		}
		verified.Close()
		return nil
	}
	if _, err := exec.LookPath("ssh-copy-id"); err != nil {
		return fmt.Errorf("ssh-copy-id is required: %w", err)
	}
	current, err := user.Current()
	if err != nil {
		return err
	}
	name, err := remoteKeyPath(current.HomeDir)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "1panel-ssh-copy-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	askpass := filepath.Join(dir, "askpass")
	if err := os.WriteFile(askpass, []byte("#!/bin/sh\nprintf '%s\\n' \"$PANEL_SSH_PASSWORD\"\n"), 0700); err != nil {
		return err
	}
	args := []string{"-i", name + ".pub", "-p", strconv.Itoa(info.Port), "-o", "StrictHostKeyChecking=accept-new", "-o", "ConnectTimeout=5", "-o", "NumberOfPasswordPrompts=1"}
	password := info.Password
	if info.AuthMode == "key" {
		bootstrap := filepath.Join(dir, "bootstrap")
		if err := os.WriteFile(bootstrap, info.PrivateKey, 0600); err != nil {
			return err
		}
		args = append(args, "-f", "-o", "IdentityFile="+bootstrap)
		password = string(info.PassPhrase)
	}
	if info.Addr == "" || strings.HasPrefix(info.Addr, "-") || strings.ContainsAny(info.Addr, "\r\n") || info.User == "" || strings.ContainsAny(info.User, "\r\n") {
		return fmt.Errorf("invalid SSH host or user")
	}
	args = append(args, "-o", "User="+info.User, info.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "ssh-copy-id", args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	command.Env = append(os.Environ(), "SSH_ASKPASS="+askpass, "SSH_ASKPASS_REQUIRE=force", "DISPLAY=1panel:0", "PANEL_SSH_PASSWORD="+password)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("ssh-copy-id failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	client, err = NewClient(probe)
	if err != nil {
		return fmt.Errorf("public key installed but passwordless verification failed: %w", err)
	}
	client.Close()
	return nil
}

func (c *SSHClient) installRemoteAuthorization(key gossh.PublicKey) error {

	session, err := c.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	timer := time.AfterFunc(30*time.Second, func() { c.Close() })
	defer timer.Stop()
	output, err := session.CombinedOutput(remoteAuthorizationScript(key))
	if err != nil {
		return fmt.Errorf("update SSH authorization failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func remoteAuthorizationScript(key gossh.PublicKey) string {
	fields := strings.Fields(string(gossh.MarshalAuthorizedKey(key)))
	return fmt.Sprintf(`set -eu
umask 077
mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"
(
flock -x 9
file="$HOME/.ssh/authorized_keys"
[ ! -L "$file" ] || { echo 'authorized_keys is a symbolic link; configure it manually' >&2; exit 1; }
[ -e "$file" ] || : > "$file"
tmp=$(mktemp "$HOME/.ssh/.1panel-authorized-XXXXXX")
trap 'rm -f "$tmp"' EXIT HUP INT TERM
awk -v kind='%s' -v blob='%s' '
{
    line = $0
    sub(/\r$/, "", line)
    sub(/^[ \t]+/, "", line)
    quoted = escaped = 0
    for (i = 1; i <= length(line); i++) {
        ch = substr(line, i, 1)
        if (escaped) { escaped = 0; continue }
        if (ch == "\\") { escaped = 1; continue }
        if (ch == "\"") { quoted = !quoted; continue }
        if (!quoted && ch ~ /[ \t]/) break
    }
    first = substr(line, 1, i - 1)
    rest = substr(line, i)
    sub(/^[ \t]+/, "", rest)
    split(rest, parts, /[ \t]+/)
    matching = (first == kind && parts[1] == blob) || (first !~ /^#/ && parts[1] == kind && parts[2] == blob)
    if (matching) found = 1
    print
}
END { if (!found) print kind " " blob " 1Panel remote SSH" }
' "$file" > "$tmp"
chmod 600 "$tmp"
mv "$tmp" "$file"
) 9> "$HOME/.ssh/.1panel-authorized.lock"
`, fields[0], fields[1])
}
