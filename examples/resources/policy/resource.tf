resource "directus_policy" "editor" {
  name        = "Content Editor"
  description = "Policy for content editors"
  icon        = "edit"

  app_access   = true
  admin_access = false
  enforce_tfa  = true
}
