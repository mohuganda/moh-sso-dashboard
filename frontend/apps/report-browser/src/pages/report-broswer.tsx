const ReportBrowser = () => {
  return (
    <div style={{ width: "100%", height: "100vh", overflow: "hidden" }}>
      <iframe
        src="https://dashboards.health.go.ug/report-browser/"
        title="Health Dashboard Report Browser"
        style={{
          width: "100%",
          height: "100%",
          border: "none",
        }}
        allowFullScreen
      />
    </div>
  );
};

export default ReportBrowser;
