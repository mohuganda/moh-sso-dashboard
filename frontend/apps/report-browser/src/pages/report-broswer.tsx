import "./report-browser.scss";

const ReportBrowser = () => {
  return (
    <div className="report-browser">
      <iframe
        src="https://dashboards.health.go.ug/report-browser/"
        title="Health Dashboard Report Browser"
        className="report-browser__frame"
        allowFullScreen
      />
    </div>
  );
};

export default ReportBrowser;
