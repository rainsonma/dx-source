import type { ImageRole } from "@/consts/image-role";

export type UploadedImage = {
  id: string;
  url: string;
  name: string;
};

export type ImageUploaderConfig = {
  role: ImageRole;
  maxFileSize?: number;
  allowedTypes?: string[];
  onUploadComplete?: (image: UploadedImage) => void;
};
