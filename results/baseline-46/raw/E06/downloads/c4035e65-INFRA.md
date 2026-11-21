# Orders infrastructure

## What `main.tf` deploys

One AWS resource, via the `aws` provider in `var.region`:

- `aws_db_instance.orders`: a single PostgreSQL RDS instance
  - `db.t3.micro`, 20 GB allocated storage
  - initial database `var.db_name`, master user `app`, password `var.db_auth`
  - `publicly_accessible = false` (no public IP)
  - `skip_final_snapshot = true`

Everything else (VPC, subnet group, security groups, parameter group, engine
version, backups, encryption, Multi-AZ) is left to AWS defaults. There is no
backend block, so state is stored locally in `terraform.tfstate`.

## Inputs

| Variable  | Type   | Sensitive | Purpose                                   |
|-----------|--------|-----------|-------------------------------------------|
| `region`  | string | no        | AWS region for the provider and instance  |
| `db_name` | string | no        | Name of the initial database               |
| `db_auth` | string | yes       | Master password for user `app`             |

None have defaults; all must be supplied (for example in `terraform.tfvars`,
`TF_VAR_*` environment variables, or `-var`).

## Operational risks

1. **Data loss on destroy**: `skip_final_snapshot = true` means `terraform destroy`
   or any replacement deletes the database with no final snapshot. Set it to
   `false` with a `final_snapshot_identifier`, and add `deletion_protection = true`.
2. **Secret in plaintext files**: `terraform.tfvars` holds the real `db_auth` and is
   not git-ignored, so it can be committed by accident. The password is also stored
   in plaintext in the state file regardless of `sensitive = true`. Keep tfvars out of
   git, use a secrets manager (or `manage_master_user_password = true`), and use an
   encrypted remote backend.
3. **Local state, no locking**: without a remote backend, state can be lost or
   diverge between operators, and concurrent runs are not locked.
4. **Default networking**: no `db_subnet_group_name` or `vpc_security_group_ids`, so
   the instance lands in the default VPC with the default security group.
   "Private" depends on that VPC's setup; pin explicit subnets and security groups.
5. **Unpinned engine and provider versions**: no `engine_version` and no
   `required_providers` block, so new applies may pick different PostgreSQL or
   provider versions. Pin both.
6. **Durability and capacity defaults**: single-AZ, no explicit backup retention,
   storage encryption not set, no storage autoscaling, and a burstable `db.t3.micro`.
   Fine for dev; for production set `multi_az`, `backup_retention_period`,
   `storage_encrypted = true`, `max_allocated_storage`, and a larger class.
7. **Changing `db_name` forces replacement**, which with risk 1 means losing data.
