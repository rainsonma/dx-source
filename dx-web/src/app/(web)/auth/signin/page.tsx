import { Suspense } from "react";
import { SignInForm } from "@/features/web/auth/components/sign-in-form";

export default function SignInPage() {
  return (
    <Suspense>
      <SignInForm />
    </Suspense>
  );
}
