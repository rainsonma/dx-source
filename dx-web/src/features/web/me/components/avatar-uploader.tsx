"use client";

import { Camera } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useImageUploader } from "@/features/com/images/hooks/use-image-uploader";
import { IMAGE_ROLES } from "@/consts/image-role";
import { getAvatarColor } from "@/lib/avatar";
import { updateAvatarAction } from "@/features/web/me/actions/me.action";

interface AvatarUploaderProps {
  userId: string;
  avatarUrl: string | null;
  displayName: string;
  onProfileChanged?: () => void;
}

/** Clickable avatar with Uppy upload overlay */
export function AvatarUploader({ userId, avatarUrl, displayName, onProfileChanged }: AvatarUploaderProps) {
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const avatarBg = getAvatarColor(userId);

  const { upload, isUploading, inputRef, handleFileChange } = useImageUploader({
    role: IMAGE_ROLES.USER_AVATAR,
    onUploadComplete: async (image) => {
      const result = await updateAvatarAction(image.url);
      if (result.success) {
        onProfileChanged?.();
      }
    },
  });

  return (
    <div className="relative cursor-pointer" onClick={upload}>
      <Avatar className="h-20 w-20">
        {avatarUrl && <AvatarImage src={avatarUrl} alt={displayName} />}
        <AvatarFallback
          className="text-2xl font-bold"
          style={{ backgroundColor: avatarBg, color: "#fff" }}
        >
          {fallbackChar}
        </AvatarFallback>
      </Avatar>

      <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition-opacity hover:opacity-100">
        <Camera className="h-6 w-6 text-white" />
      </div>

      {isUploading && (
        <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/50">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent" />
        </div>
      )}

      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png"
        className="hidden"
        onChange={handleFileChange}
      />
    </div>
  );
}
