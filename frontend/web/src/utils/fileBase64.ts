/** 将 File / Blob 转为纯 Base64（不含 data: 前缀） */
export function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      const idx = result.indexOf(',')
      resolve(idx >= 0 ? result.slice(idx + 1) : result)
    }
    reader.onerror = () => reject(reader.error ?? new Error('读取文件失败'))
    reader.readAsDataURL(blob)
  })
}

export function fileToBase64(file: File): Promise<string> {
  return blobToBase64(file)
}

/** 浏览器录音多为 webm，百炼要求 wav/mpeg/mp4，解码后转 WAV */
export async function blobToWavBase64(blob: Blob): Promise<{ base64: string, mime: string }> {
  const ctx = new AudioContext()
  try {
    const audioBuffer = await ctx.decodeAudioData(await blob.arrayBuffer())
    const wavBlob = encodeWavBlob(audioBuffer)
    const base64 = await blobToBase64(wavBlob)
    return { base64, mime: 'audio/wav' }
  }
  finally {
    await ctx.close()
  }
}

/** 录音/上传样本编码：webm 转 wav，其余保持原格式 */
export async function encodeSampleForClone(blob: Blob, mime?: string): Promise<{ base64: string, mime: string }> {
  const m = (mime || blob.type || '').toLowerCase()
  if (m.includes('webm') || m.includes('ogg')) {
    return blobToWavBase64(blob)
  }
  return { base64: await blobToBase64(blob), mime: m || 'audio/wav' }
}

function encodeWavBlob(audioBuffer: AudioBuffer): Blob {
  const numChannels = Math.min(audioBuffer.numberOfChannels, 1)
  const sampleRate = audioBuffer.sampleRate
  const samples = audioBuffer.getChannelData(0)
  const buffer = new ArrayBuffer(44 + samples.length * 2)
  const view = new DataView(buffer)

  const writeString = (offset: number, str: string) => {
    for (let i = 0; i < str.length; i++)
      view.setUint8(offset + i, str.charCodeAt(i))
  }

  writeString(0, 'RIFF')
  view.setUint32(4, 36 + samples.length * 2, true)
  writeString(8, 'WAVE')
  writeString(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, numChannels, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, sampleRate * numChannels * 2, true)
  view.setUint16(32, numChannels * 2, true)
  view.setUint16(34, 16, true)
  writeString(36, 'data')
  view.setUint32(40, samples.length * 2, true)

  let offset = 44
  for (let i = 0; i < samples.length; i++, offset += 2) {
    const s = Math.max(-1, Math.min(1, samples[i]))
    view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7FFF, true)
  }

  return new Blob([buffer], { type: 'audio/wav' })
}
