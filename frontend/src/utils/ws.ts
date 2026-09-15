// ws.ts WebSocket 连接管理（消息实时推送）。
export function createMessageSocket(token: string): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = window.location.host
  return new WebSocket(`${protocol}://${host}/api/v1/ws?token=${encodeURIComponent(token)}`)
}
