import React, { useRef } from 'react';
import { VncScreen } from 'react-vnc';

export default function DynamicVNCViewer({ url, password }) {
  const vncRef = useRef();

  return (
    <VncScreen
      ref={vncRef}
      url={url}
      scaleViewport
      background="#000"
      onCredentialsRequired={() => {
        // Called when server requests credentials
        if (vncRef.current) {
          vncRef.current.sendCredentials({ password });
        }
      }}
      style={{
        width: '75vw',
        height: '75vh',
      }}
      onConnect={() => console.log("VNC connected")}
      onDisconnect={() => console.log("VNC disconnected")}
    />
  );
}
