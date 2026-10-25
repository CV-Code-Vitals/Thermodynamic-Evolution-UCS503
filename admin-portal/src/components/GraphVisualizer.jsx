import React, { useCallback, useEffect, useRef } from 'react';

const graphURL = `${import.meta.env.BASE_URL}graph-visualizer/index.html?embedded=true`;

export default function GraphVisualizer({ report }) {
  const frameRef = useRef(null);
  const readyRef = useRef(false);

  const sendReport = useCallback(() => {
    const target = frameRef.current?.contentWindow;
    if (!target) return;
    target.postMessage({ type: 'ucs503:report', report: report || { file_reports: [] } }, window.location.origin);
  }, [report]);

  useEffect(() => {
    if (readyRef.current) sendReport();
  }, [report, sendReport]);

  return (
    <iframe
      ref={frameRef}
      className="graph-visualizer-frame"
      src={graphURL}
      title="Repository graph visualizer"
      onLoad={() => {
        readyRef.current = true;
        sendReport();
      }}
      onError={() => { readyRef.current = false; }}
    />
  );
}
