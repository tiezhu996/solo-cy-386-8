<template>
  <el-card class="msg-wrap">
    <div class="layout">
      <div class="conv-list">
        <div class="conv-head">会话列表</div>
        <div v-for="c in conversations" :key="c.peer_id" class="conv" :class="{ active: c.peer_id === peerId }" @click="openConversation(c)">
          <div class="conv-name">{{ c.peer_name }}</div>
          <div class="conv-last">{{ c.last_content }}</div>
          <el-badge v-if="c.unread_count > 0" :value="c.unread_count" />
        </div>
        <EmptyState v-if="!conversations.length" description="暂无会话" />
      </div>
      <div class="chat">
        <template v-if="peerId">
          <div class="chat-head">
            与 {{ peerName }} 的聊天
            <router-link v-if="peerProductId" :to="`/products/${peerProductId}`" class="prod-link">查看商品</router-link>
          </div>
          <div ref="chatBox" class="chat-body">
            <div v-for="m in messages" :key="m.id" class="msg-row" :class="{ mine: m.sender_id === userStore.user?.id }">
              <div class="bubble">{{ m.content }}</div>
              <div class="time">{{ m.created_at }}</div>
            </div>
          </div>
          <div class="chat-input">
            <el-input v-model="content" placeholder="输入消息..." @keyup.enter="send" />
            <el-button type="primary" @click="send">发送</el-button>
          </div>
        </template>
        <div v-else class="placeholder">选择一个会话开始聊天</div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as messageApi from '../api/message'
import { useUserStore } from '../stores/userStore'
import { useMessageStore } from '../stores/messageStore'
import { createMessageSocket } from '../utils/ws'
import EmptyState from '../components/EmptyState.vue'

const route = useRoute()
const userStore = useUserStore()
const messageStore = useMessageStore()
const conversations = ref<any[]>([])
const messages = ref<any[]>([])
const peerId = ref<number | null>(null)
const peerName = ref('')
const peerProductId = ref<number | null>(null)
const content = ref('')
const chatBox = ref<HTMLElement | null>(null)
let ws: WebSocket | null = null

onMounted(async () => {
  await loadConversations()
  const qPeer = Number(route.query.peer_id)
  const qProduct = Number(route.query.product_id)
  if (qPeer) {
    peerId.value = qPeer
    peerProductId.value = qProduct || null
    await openPeer(qPeer, qProduct || 0)
  }
  connectWs()
})

watch(peerId, async () => {
  if (peerId.value) {
    await messageApi.markRead(peerId.value)
    messageStore.loadUnread()
  }
})

function connectWs() {
  const token = localStorage.getItem('marketpal_token')
  if (!token) return
  try {
    ws = createMessageSocket(token)
    ws.onmessage = () => {
      messageStore.onWsMessage()
      if (peerId.value) {
        loadMessages(peerId.value)
      }
      loadConversations()
    }
  } catch {
    // WebSocket 不可用时降级为手动刷新。
  }
}

async function loadConversations() {
  const res: any = await messageApi.listConversations()
  conversations.value = res.data || []
}

async function openConversation(c: any) {
  peerId.value = c.peer_id
  peerName.value = c.peer_name
  peerProductId.value = c.product_id || null
  await openPeer(c.peer_id, c.product_id || 0)
}

async function openPeer(pid: number, productId: number) {
  if (!peerName.value) {
    const c = conversations.value.find((x) => x.peer_id === pid)
    peerName.value = c?.peer_name || `用户${pid}`
    peerProductId.value = c?.product_id || productId || null
  }
  await loadMessages(pid)
}

async function loadMessages(pid: number) {
  const res: any = await messageApi.listConversation(pid, { page: 1, page_size: 50 })
  messages.value = (res.data.list || []).slice().reverse()
  await nextTick()
  if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
}

async function send() {
  if (!content.value.trim() || !peerId.value) return
  await messageApi.sendMessage({ receiver_id: peerId.value, product_id: peerProductId.value || undefined, content: content.value.trim() })
  content.value = ''
  await loadMessages(peerId.value)
  await loadConversations()
  ElMessage.success('消息已发送')
}
</script>

<style scoped>
.msg-wrap {
  height: 640px;
}
.layout {
  display: flex;
  height: 100%;
}
.conv-list {
  width: 280px;
  border-right: 1px solid #ebeef5;
  overflow-y: auto;
}
.conv-head {
  padding: 12px;
  font-weight: 600;
  border-bottom: 1px solid #f0f2f5;
}
.conv {
  padding: 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.conv.active {
  background: #ecf5ff;
}
.conv-name {
  font-weight: 600;
}
.conv-last {
  color: #909399;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}
.chat {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.chat-head {
  padding: 12px;
  font-weight: 600;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  justify-content: space-between;
}
.prod-link {
  font-size: 13px;
  color: #409eff;
}
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  background: #fafafa;
}
.msg-row {
  margin-bottom: 12px;
}
.msg-row.mine {
  text-align: right;
}
.bubble {
  display: inline-block;
  max-width: 70%;
  background: #fff;
  border-radius: 8px;
  padding: 8px 12px;
  border: 1px solid #ebeef5;
  text-align: left;
}
.mine .bubble {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}
.time {
  font-size: 12px;
  color: #c0c4cc;
  margin-top: 4px;
}
.chat-input {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-top: 1px solid #ebeef5;
}
.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
  height: 100%;
}
</style>
