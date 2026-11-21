# Orders infrastructure

## What `main.tf` deploys

One AWS resource, `aws_db_instance.orders`: a single PostgreSQL RDS instance.

| Setting | Value |
| --- | --- |
| Engine | `postgres` (version not pinned; AWS default at create time) |
| Instance class | `db.t3.micro` (burstable, 2 vCPU / 1 GiB) |
| Storage | 20 GiB, no autoscaling (`max_allocated_storage` unset) |
| Database name | `var.db_name` |
| Master user | `app`, password from `var.db_auth` |
| Network | `publicly_accessible = false`; no subnet group or security groups set, so it lands in the account's default VPC with the default security group |
| Final snapshot | Skipped on destroy (`skip_final_snapshot = true`) |

The AWS provider is configured with `var.region`. There is no `terraform` block, so no backend (state is local) and no provider version pin.

## Inputs

| Variable | Type | Sensitive | Purpose |
| --- | --- | --- | --- |
| `region` | string | no | AWS region for the provider and the instance |
| `db_name` | string | no | Name of the initial database created in the instance |
| `db_auth` | string | yes | Master password for user `app` |

None has a default; all three must be supplied (currently via an untracked `terraform.tfvars`).

## Operational risks

1. **Plaintext secret in `terraform.tfvars`.** The file holds the master password and is not in `.gitignore`; it can be committed by accident. Add it to `.gitignore` and move the password to a secrets manager (e.g. `manage_master_user_password = true` or a `TF_VAR_db_auth` env var).
2. **Password in state.** `sensitive = true` only hides it from CLI output; it is stored in plaintext in `terraform.tfstate`, which is local and unencrypted because no backend is configured. Use a remote, encrypted, access-controlled backend with locking.
3. **Data loss on destroy.** `skip_final_snapshot = true` and no `deletion_protection` mean `terraform destroy` or a replacing change deletes the database with no final snapshot.
4. **Backups not set explicitly.** `backup_retention_period` is left to the default; set it deliberately and confirm point-in-time recovery meets requirements.
5. **No encryption at rest.** `storage_encrypted` is unset (defaults to false); it cannot be turned on later without a snapshot-and-restore replacement.
6. **Single AZ.** No `multi_az`; an AZ outage or maintenance causes downtime.
7. **Default networking.** Without `db_subnet_group_name` and `vpc_security_group_ids`, the instance uses the default VPC and default security group, which may allow broader access than intended. "Private" depends entirely on that default setup.
8. **Unpinned versions.** No `engine_version` and no provider `required_providers` constraint; re-runs on another machine or later date may plan differently or trigger upgrades.
9. **Capacity.** `db.t3.micro` with 20 GiB and no storage autoscaling is a dev-sized footprint; CPU credits and disk can run out under production load.
10. **No `identifier`.** AWS/Terraform generates a random instance name, which makes the instance harder to find and reference operationally.
