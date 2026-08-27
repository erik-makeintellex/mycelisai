import { OutputCanvas } from "@/components/soma/OutputCanvas";

interface OutputCanvasSearchParams {
  source?: string | string[];
  label?: string | string[];
  path?: string | string[];
  return_to?: string | string[];
  proof?: string | string[];
  team_id?: string | string[];
  run_id?: string | string[];
  work_item_id?: string | string[];
  output_id?: string | string[];
  digest?: string | string[];
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
        teamId={first(params?.team_id)}
        runId={first(params?.run_id)}
        workItemId={first(params?.work_item_id)}
        outputId={first(params?.output_id)}
        contentDigest={first(params?.digest)}
      />
    </div>
  );
}
