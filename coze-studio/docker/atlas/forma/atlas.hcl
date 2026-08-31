variable "forma_migration_dir" {
  default = "file://migrations"
}

env "forma-local" {
  url = getenv("FORMA_ATLAS_URL")
  dev = "docker://mysql/8/forma"
  migration {
    dir = var.forma_migration_dir
  }
}
