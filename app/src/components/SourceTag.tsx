import type { Source } from "../types";

const LABEL: Record<Source, string> = {
  local: "local",
  peer: "peer",
  indexer: "indexer",
};

const TITLE: Record<Source, string> = {
  local: "Answered from your own shard — always available, offline-safe.",
  peer: "Reached over the mesh from another volunteer node.",
  indexer: "An opt-in global indexer — adds reach, never authority.",
};

export function SourceTag({ source }: { source: Source }) {
  return (
    <span className={`source-tag source-tag--${source}`} title={TITLE[source]}>
      <span className="source-tag-dot" aria-hidden="true" />
      {LABEL[source]}
    </span>
  );
}
