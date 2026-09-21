/**
 * audioProcessor.ts
 * Captures microphone input via Web Audio API, downsamples to 16kHz,
 * converts to PCM16, and fires an onData callback.
 */

export class AudioProcessor {
  private stream: MediaStream | null = null;
  private audioContext: AudioContext | null = null;
  private processor: ScriptProcessorNode | null = null;
  private source: MediaStreamAudioSourceNode | null = null;

  private onData: (data: Int16Array) => void;

  constructor(onData: (data: Int16Array) => void) {
    this.onData = onData;
  }

  async start() {
    try {
      this.stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 16000,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });
    } catch (err) {
      console.error("Microphone access denied or error", err);
      throw err;
    }

    // Use standard AudioContext
    const AudioContextClass =
      window.AudioContext || (window as any).webkitAudioContext;
    this.audioContext = new AudioContextClass({ sampleRate: 16000 });

    this.source = this.audioContext.createMediaStreamSource(this.stream);

    // Create a processor node. Buffer size 4096 is standard.
    // We request 1 input channel, 1 output channel.
    this.processor = this.audioContext.createScriptProcessor(4096, 1, 1);

    this.processor.onaudioprocess = (e) => {
      const inputBuffer = e.inputBuffer.getChannelData(0);
      // Downconvert Float32 (from Web Audio) to Int16 (PCM16)
      const pcm16 = new Int16Array(inputBuffer.length);
      for (let i = 0; i < inputBuffer.length; i++) {
        let s = Math.max(-1, Math.min(1, inputBuffer[i]));
        pcm16[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
      }
      this.onData(pcm16);
    };

    // Connect the graph
    this.source.connect(this.processor);
    this.processor.connect(this.audioContext.destination);
  }

  stop() {
    if (this.processor) {
      this.processor.disconnect();
      this.processor.onaudioprocess = null;
      this.processor = null;
    }
    if (this.source) {
      this.source.disconnect();
      this.source = null;
    }
    if (this.audioContext && this.audioContext.state !== "closed") {
      this.audioContext.close();
      this.audioContext = null;
    }
    if (this.stream) {
      this.stream.getTracks().forEach((t) => t.stop());
      this.stream = null;
    }
  }
}
