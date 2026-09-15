// messageStore.ts 私信状态管理（未读数、实时消息回调）。
import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as messageApi from '../api/message'

export const useMessageStore = defineStore('message', () => {
  const unread = ref(0)

  async function loadUnread() {
    try {
      const res: any = await messageApi.unreadCount()
      unread.value = Number(res.data.unread || 0)
    } catch {
      unread.value = 0
    }
  }

  function onWsMessage() {
    loadUnread()
  }

  return { unread, loadUnread, onWsMessage }
})
