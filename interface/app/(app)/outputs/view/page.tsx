import { OutputCanvas } from "@/components/soma/OutputCanvas";

interface OutputCanvasSearchParams {
  source?: string | string[];
  label?: string | string[];
  path?: string | string[];
  return_to?: string | string[];
  proof?: string | string[];
}

function first(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}

export default async function OutputCanvasPage({
  searchParams,
}: {
  searchParams?: Promise<OutputCanvasSearchParams>;
}) {
  const params = await searchParams;
  return (
    <div className="h-full min-h-0 overflow-hidden">
      <OutputCanvas
        source={first(params?.source)}
        label={first(params?.label)}
        storagePath={first(params?.path)}
        returnTo={first(params?.return_to)}
        proofArtifactId={first(params?.proof)}
      />
    </div>
  );
}
