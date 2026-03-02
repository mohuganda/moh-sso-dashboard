export type StorageLocation = {
  id: string;
  code: string;
  name: string;
  provider: "local" | "s3" | "minio" | "nfs";
  base_uri: string;
  is_active: boolean;
  created_at: string;
};
