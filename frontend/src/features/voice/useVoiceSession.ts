import { useState, useRef, useEffect, useCallback } from "react";
import { AudioProcessor } from "./audioProcessor";
import { WSControlMessage } from "./types";
import { getWebSocketURL } from "./api";

type ConnectionStatus = "DISCONNECTED" | "CONNECTING" | "CONNECTED" | "LISTENING" | "PAUSED" | "ERROR";

export function useVoiceSession(sessionId?: string) {
  const [status, setStatus] = useState<ConnectionStatus>("DISCONNECTED");
  const [error, setError] = useState<string | null>(null);

  const ws = useRef<WebSocket | null>(null);
  const audio = useRef<AudioProcessor | null>(null);
  const sequence = useRef<number>(0);

  const connect = useCallback(() => {
    if (!sessionId) return;
    setStatus("CONNECTING");

    const url = getWebSocketURL(sessionId);
    const socket = new WebSocket(url);
    socket.binaryType = "arraybuffer";

    socket.onopen = () => {
      // Audio processor will start when we send START and get READY
      setStatus("CONNECTED");
    };

    socket.onmessage = (event) => {
      if (typeof event.data === "string") {
        try {
          const msg = JSON.parse(event.data) as WSControlMessage;
          switch (msg.type) {
            case "CONNECTED":
              setStatus("CONNECTED");
              break;
            case "READY":
              setStatus("LISTENING");
              startAudioCapture();
              break;
            case "STOPPED":
              disconnect();
              break;
            case "ERROR":
              setError(msg.message || "Unknown error");
              setStatus("ERROR");
              stopAudioCapture();
              break;
            case "WARNING":
              console.warn("WS Warning:", msg.message);
              break;
          }
        } catch (e) {
          console.error("Failed to parse WS text frame", e);
        }
      }
    };

    socket.onerror = (e) => {
      console.error("WS error", e);
      setError("WebSocket error occurred");
      setStatus("ERROR");
    };

    socket.onclose = () => {
      setStatus("DISCONNECTED");
      stopAudioCapture();
    };

    ws.current = socket;
  }, [sessionId]);

  const startAudioCapture = async () => {
    if (audio.current) return;
    const processor = new AudioProcessor((pcm16) => {
      sendAudioChunk(pcm16);
    });
    try {
      await processor.start();
      audio.current = processor;
    } catch (err: any) {
      setError(err.message);
    }
  };

  const stopAudioCapture = () => {
    if (audio.current) {
      audio.current.stop();
      audio.current = null;
    }
  };

  const sendAudioChunk = (pcm16: Int16Array) => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) return;
    if (status !== "LISTENING") return; // don't send if paused

    const seq = sequence.current++;
    const ts = Date.now(); // absolute MS since epoch

    // Protocol: Sequence (uint32) + TimestampMs (uint64) + PCM (int16 array)
    const headerSize = 12; // 4 + 8
    const byteLength = headerSize + pcm16.byteLength;
    const buffer = new ArrayBuffer(byteLength);
    const view = new DataView(buffer);

    // Write Sequence (uint32, little-endian)
    view.setUint32(0, seq, true);

    // Write TimestampMs (uint64, little-endian)
    // Note: DataView.setBigUint64 is available in modern browsers
    view.setBigUint64(4, BigInt(ts), true);

    // Write Audio
    const audioView = new Int16Array(buffer, headerSize);
    audioView.set(pcm16);

    ws.current.send(buffer);
  };

  const sendControl = (type: string, message?: string) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type, message }));
    }
  };

  const startListening = () => sendControl("START");
  const pauseListening = () => {
    sendControl("PAUSE");
    setStatus("PAUSED"); // optimistic
  };
  const resumeListening = () => {
    sendControl("RESUME");
    setStatus("LISTENING");
  };
  const stopListening = () => sendControl("STOP");

  const disconnect = useCallback(() => {
    stopAudioCapture();
    if (ws.current) {
      ws.current.close();
      ws.current = null;
    }
    setStatus("DISCONNECTED");
  }, []);

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    status,
    error,
    connect,
    disconnect,
    startListening,
    pauseListening,
    resumeListening,
    stopListening,
  };
}
