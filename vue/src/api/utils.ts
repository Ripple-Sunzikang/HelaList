export function getToken() {
  return localStorage.getItem('token') || ''
}

export async function download(url: string, onProgress: (progress: number) => void) {
  const token = getToken()
  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(url, { headers })
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const contentLength = response.headers.get('content-length')
  const total = contentLength ? parseInt(contentLength, 10) : 0
  let loaded = 0

  const reader = response.body!.getReader()
  const stream = new ReadableStream({
    start(controller) {
      function push() {
        reader.read().then(({ done, value }) => {
          if (done) {
            controller.close()
            return
          }
          loaded += value.length
          if (total) {
            onProgress(Math.round((loaded / total) * 100))
          }
          controller.enqueue(value)
          push()
        })
      }
      push()
    },
  })

  return new Response(stream)
}
