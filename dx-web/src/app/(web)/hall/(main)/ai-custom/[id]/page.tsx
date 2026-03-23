import { CourseDetailContent } from "@/features/web/ai-custom/components/course-detail-content";

export default async function CourseGameDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <CourseDetailContent id={id} />
    </div>
  );
}
