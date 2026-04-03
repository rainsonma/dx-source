"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Uppy from "@uppy/core";
import XHRUpload from "@uppy/xhr-upload";

import type { ImageRole } from "@/consts/image-role";
import type { UploadedImage } from "@/features/com/images/types/image.types";

type UseImageUploaderOptions = {
  role: ImageRole;
  maxFileSize?: number;
  allowedTypes?: string[];
  onUploadComplete?: (image: UploadedImage) => void;
};

const DEFAULT_MAX_SIZE = 2 * 1024 * 1024;
const DEFAULT_TYPES = ["image/jpeg", "image/png"];
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

export function useImageUploader({
  role,
  maxFileSize = DEFAULT_MAX_SIZE,
  allowedTypes = DEFAULT_TYPES,
  onUploadComplete,
}: UseImageUploaderOptions) {
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [image, setImage] = useState<UploadedImage | null>(null);
  const [error, setError] = useState<string | null>(null);

  const uppyRef = useRef<Uppy<{ role: string }, UploadedImage> | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);
  const onUploadCompleteRef = useRef(onUploadComplete);
  useEffect(() => { onUploadCompleteRef.current = onUploadComplete; });

  useEffect(() => {
    const uppy = new Uppy<{ role: string }, UploadedImage>({
      restrictions: {
        maxFileSize,
        allowedFileTypes: allowedTypes,
        maxNumberOfFiles: 1,
      },
      autoProceed: true,
    });

    uppy.use(XHRUpload, {
      endpoint: `${API_URL}/api/uploads/images`,
      fieldName: "file",
      formData: true,
      withCredentials: true,
      allowedMetaFields: ["role"],
      getResponseData(xhr: XMLHttpRequest) {
        const parsed = JSON.parse(xhr.responseText);
        // Go API wraps response in {code, message, data} envelope
        return parsed.data ?? parsed;
      },
    });

    uppy.setMeta({ role });

    uppy.on("upload", () => {
      setIsUploading(true);
      setError(null);
      setProgress(0);
    });

    uppy.on("progress", (p) => {
      setProgress(p);
    });

    uppy.on("upload-success", (_file, response) => {
      const uploaded = response.body as UploadedImage;
      setImage(uploaded);
      setIsUploading(false);
      setProgress(100);
      onUploadCompleteRef.current?.(uploaded);
    });

    uppy.on("upload-error", (_file, err) => {
      setError(err?.message || "上传失败");
      setIsUploading(false);
      setProgress(0);
    });

    uppy.on("restriction-failed", (_file, err) => {
      setError(err?.message || "文件不符合要求");
    });

    uppyRef.current = uppy;

    return () => {
      uppy.clear();
      uppy.destroy();
    };
  }, [role, maxFileSize, allowedTypes]);

  const upload = useCallback(() => {
    inputRef.current?.click();
  }, []);

  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file || !uppyRef.current) return;

      // Clear previous files
      uppyRef.current.clear();

      try {
        uppyRef.current.addFile({
          name: file.name,
          type: file.type,
          data: file,
          source: "local",
        });
      } catch {
        // Restriction errors are handled via the restriction-failed event
      }

      // Reset input so same file can be selected again
      e.target.value = "";
    },
    [],
  );

  const reset = useCallback(() => {
    uppyRef.current?.clear();
    setImage(null);
    setError(null);
    setProgress(0);
    setIsUploading(false);
  }, []);

  return {
    upload,
    reset,
    isUploading,
    progress,
    image,
    error,
    inputRef,
    handleFileChange,
  };
}
