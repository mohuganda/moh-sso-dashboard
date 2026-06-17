import "./BuildMeta.scss";

type BuildMetaProps = {
  version?: string;
  buildTime?: string;
};

export function BuildMeta({ version, buildTime }: BuildMetaProps) {
  return (
    <div className="moh-build-meta">
      v{version ?? "0.0.0"}
      {buildTime && (
        <>
          {" · "}
          built {new Date(buildTime).toLocaleString()}
        </>
      )}
    </div>
  );
}
