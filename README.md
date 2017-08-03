# dgraph_helper

This should only be used on Ubuntu 15.04 or higher (or any linux with systemd).

This is a work in progress.

### Usage

  1. Download or install dgraph_helper
  2. Run dgraph_helper
  3. Follow the prompts.

Dgraph will be installed, configured, and started as a service under systemd.

After installing dgraph type: 

  + `systemctl status dgraph` to see dgraph's status
  + `systemctl stop dgraph` to stop dgraph
  + `systemctl restart dgraph` to restart dgraph
  + to run dgraph in the current terminal (not as a service) stop the service with `systemctl stop dgraph` and then run `dgraph --config=/var/lib/dgraph/config.yaml` (or `--config=<your_config_dot_yaml_here>`)


### download and run it.


```
curl -L https://github.com/elbow-jason/dgraph_helper/releases/download/0.1.0/dgraph_helper_v0.1.0_linux_amd64 -o dgraph_helper && chmod +x dgraph_helper && ./dgraph_helper
```
