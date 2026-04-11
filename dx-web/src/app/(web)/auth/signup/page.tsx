import { cookies } from "next/headers";

import { SignUpForm } from "@/features/web/auth/components/sign-up-form";

export default async function SignUpPage() {
  const cookieStore = await cookies();
  const hasInviteRef = Boolean(cookieStore.get("ref")?.value);
  return <SignUpForm hasInviteRef={hasInviteRef} />;
}
