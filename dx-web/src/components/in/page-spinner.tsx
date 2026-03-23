import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

type PageSpinnerProps = {
  size?: "sm" | "md" | "lg"
  className?: string
}

const sizeMap = { sm: "h-4 w-4", md: "h-5 w-5", lg: "h-6 w-6" }
const paddingMap = { sm: "py-4", md: "py-12", lg: "py-20" }

export function PageSpinner({ size = "md", className }: PageSpinnerProps) {
  return (
    <div className={cn("flex items-center justify-center", paddingMap[size], className)}>
      <Loader2 className={cn("animate-spin text-muted-foreground", sizeMap[size])} />
    </div>
  )
}
