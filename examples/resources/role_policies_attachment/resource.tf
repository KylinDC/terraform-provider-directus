resource "directus_policy" "editor" {
  name        = "Content Editor"
  description = "Policy for content editors"
  app_access  = true
}

resource "directus_role" "content_team" {
  name        = "Content Team"
  description = "Content editors and writers"
}

resource "directus_role_policies_attachment" "content_team_policies" {
  role_id    = directus_role.content_team.id
  policy_ids = [directus_policy.editor.id]
}
