# Orders infrastructure

`main.tf` deploys a single AWS RDS PostgreSQL instance (`aws_db_instance.orders`) for the orders service. Nothing else is managed here: no VPC, subnets, DB subnet group, security groups, parameter group, KMS key, or remote state backend.

## What gets created

| Setting | Value | Notes |
|---|---|---|
| Engine | `postgres` | No `engine_version`, so AWS picks its current default |
| Instance class | `db.t3.micro` | Burstable, 2 vCPU / 1 GiB; dev/test sized |
| Storage | 20 GiB | Default storage type; no autoscaling (`max_allocated_storage` unset) |
| Master user | `app` | Hard-coded |
| Master password | `var.db_auth` | |
| Network | `publicly_accessible = false` | Placed in the account's default VPC/subnets and default security group |
| Deletion | `skip_final_snapshot = true` | No snapshot is taken on destroy |

## Inputs

| Variable | Type | Sensitive | Purpose | Value in `terraform.tfvars` |
|---|---|---|---|---|
| `region` | string | no | AWS region for the provider | `us-east-1` |
| `db_name` | string | no | Initial database created in the instance | `orders` |
| `db_auth` | string | yes | Master password for user `app` | set (not reproduced here) |

None of the variables has a default, so all three must be supplied (tfvars file, `-var`, or `TF_VAR_*`).

## Operational risks

1. **Data loss on destroy or replacement.** `skip_final_snapshot = true` and no `deletion_protection` mean `terraform destroy`, or any change that forces replacement (e.g. changing `db_name`), deletes the database with no final snapshot. Set `deletion_protection = true` and `skip_final_snapshot = false` with a `final_snapshot_identifier` for production.
2. **Password in plaintext.** `terraform.tfvars` holds the real master password and is not git-ignored; it must not be committed. The password is also stored in plaintext in Terraform state, and there is no backend block, so state is a local `terraform.tfstate` file. Prefer `manage_master_user_password = true` (Secrets Manager) or inject via `TF_VAR_db_auth`, and use an encrypted remote backend with locking.
3. **No encryption at rest.** `storage_encrypted` is unset (defaults to false). Enabling it later forces instance replacement, so decide before the first apply.
4. **Implicit networking.** With no `db_subnet_group_name` or `vpc_security_group_ids`, the instance uses the default VPC and default security group. "Private" depends on that default setup; access rules are not under Terraform control.
5. **Unpinned versions.** No `engine_version` and no `required_providers`/`required_version` block, so the Postgres major version and the AWS provider version can drift between runs and environments.
6. **Availability and backups.** Single-AZ (`multi_az` unset), default 1-day backup retention (`backup_retention_period` unset, AWS default for the API is 1 day), and no maintenance or backup windows. An AZ outage or maintenance causes downtime.
7. **Capacity.** `db.t3.micro` with 1 GiB RAM and burst credits, plus fixed 20 GiB storage with no autoscaling; a full disk makes the instance unavailable. No monitoring, alarms, or Performance Insights are configured.
8. **Fixed master username.** `app` is hard-coded and is the master (superuser-level) account; the application should use a separate least-privilege role.
