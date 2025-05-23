CREATE TABLE IF NOT EXISTS "l_g_openstack_image_to_cloud_profile" (
    "openstack_image_id" UUID NOT NULL,
    "cloud_profile_id" UUID NOT NULL,

    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("openstack_image_id") REFERENCES "g_cloud_profile_openstack_image" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("cloud_profile_id") REFERENCES "g_cloud_profile" ("id") ON DELETE CASCADE,
    CONSTRAINT "l_g_openstack_image_to_cloud_profile_key" UNIQUE ("openstack_image_id", "cloud_profile_id")
);
