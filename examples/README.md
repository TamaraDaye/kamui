# Examples

Two files, two responsibilities:

| File | Kind | Answers |
|---|---|---|
| `machine.yaml` | `Machine` | How do we obtain a machine? |
| `server.yaml` | `Server` | What must the machine look like? |

The intended workflow is:

```bash
# Create the VM, wait for SSH, and emit its target address.
kamui provision -f examples/machine.yaml

# Copy the emitted address into examples/server.yaml, then converge it.
kamui apply -f examples/server.yaml
```

`server.yaml` also works against any existing Linux VM. Only the `spec.target`
section needs to change.

## Before running

1. Replace the placeholder `authorized_keys` value with your real public key:

   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```

2. Update `spec.target.host` in `server.yaml` to the provisioned or existing
   VM address.

3. Make sure the target allows passwordless sudo for the configured SSH user.
   The `v0.1.0` transport uses `sudo -n` only.

4. Keep private keys out of these files. Both examples reference a key path,
   not key material.

These examples use the four must-ship resources: `user`, `package`, `file`, and
`service`. Group, directory, and firewall examples can be added after those
executors land.
