"use client";

import { CloudUpload, X, Loader2 } from "lucide-react";

import type { ImageRole } from "@/consts/image-role";
import { useImageUploader } from "@/features/com/images/hooks/use-image-uploader";

type ImageUploaderProps = {
  role: ImageRole;
  onImageChange?: (imageId: string | null) => void;
  className?: string;
};

export function ImageUploader({
  role,
  onImageChange,
  className,
}: ImageUploaderProps) {
  const {
    upload,
    reset,
    isUploading,
    progress,
    image,
    error,
    inputRef,
    handleFileChange,
  } = useImageUploader({
    role,
    onUploadComplete: (img) => onImageChange?.(img.id),
  });

  return (
    <div className={className}>
      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png"
        className="hidden"
        onChange={handleFileChange}
      />

      {image ? (
        /* Preview after upload */
        <div className="relative flex items-center gap-3 rounded-[10px] border border-border bg-muted p-3">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={image.url}
            alt={image.name}
            className="h-16 w-16 rounded-lg object-cover"
          />
          <div className="flex flex-1 flex-col gap-0.5 overflow-hidden">
            <span className="truncate text-sm font-medium text-foreground">
              {image.name}
            </span>
            <span className="text-xs text-emerald-600">上传成功</span>
          </div>
          <button
            type="button"
            onClick={() => {
              reset();
              onImageChange?.(null);
            }}
            className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-border hover:bg-accent"
          >
            <X className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
      ) : (
        /* Upload drop zone */
        <button
          type="button"
          onClick={upload}
          disabled={isUploading}
          className="flex w-full flex-col items-center justify-center gap-1.5 rounded-[10px] border border-dashed border-border bg-muted py-8 transition-colors hover:border-teal-400 hover:bg-teal-50/30 disabled:pointer-events-none disabled:opacity-60"
        >
          {isUploading ? (
            <>
              <Loader2 className="h-7 w-7 animate-spin text-teal-500" />
              <span className="text-[13px] font-semibold text-teal-600">
                上传中 {progress}%
              </span>
            </>
          ) : (
            <>
              <CloudUpload className="h-7 w-7 text-muted-foreground" />
              <span className="text-[13px] font-semibold text-muted-foreground">
                点击上传
              </span>
              <span className="text-[11px] text-muted-foreground/60">JPG / PNG</span>
            </>
          )}
        </button>
      )}

      {error && <p className="mt-1.5 text-xs text-red-500">{error}</p>}
    </div>
  );
}
