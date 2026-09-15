// cartStore.ts 购物车状态管理（数量角标）。
import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as cartApi from '../api/cart'

export const useCartStore = defineStore('cart', () => {
  const count = ref(0)
  const total = ref(0)
  const items = ref<any[]>([])

  async function load() {
    const res: any = await cartApi.getCart()
    items.value = res.data.items || []
    total.value = Number(res.data.total || 0)
    count.value = items.value.reduce((s, it) => s + it.quantity, 0)
  }

  async function add(productId: number, quantity = 1) {
    await cartApi.addToCart(productId, quantity)
    await load()
  }

  async function remove(id: number) {
    await cartApi.removeCartItem(id)
    await load()
  }

  async function update(id: number, data: { quantity?: number; selected?: boolean }) {
    await cartApi.updateCartItem(id, data)
    await load()
  }

  function reset() {
    count.value = 0
    total.value = 0
    items.value = []
  }

  return { count, total, items, load, add, remove, update, reset }
})
