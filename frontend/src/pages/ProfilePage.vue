<template>
  <el-card v-loading="loading">
    <div class="head">
      <el-avatar :size="64">{{ (userStore.user?.nickname || 'U').slice(0, 1) }}</el-avatar>
      <div>
        <h2>{{ userStore.user?.nickname }}</h2>
        <div class="credit">信用积分：{{ userStore.user?.credit_score }} · 角色：{{ formatRole(userStore.user?.role || 'user') }}</div>
      </div>
      <el-button class="edit-btn" @click="editProfile">编辑资料</el-button>
    </div>

    <el-tabs v-model="tab">
      <el-tab-pane label="我的发布" name="products">
        <div v-loading="subLoading">
          <el-table :data="myProducts" v-if="myProducts.length">
            <el-table-column prop="title" label="商品" />
            <el-table-column label="售价" width="120"><template #default="{ row }">¥{{ formatPrice(row.price) }}</template></el-table-column>
            <el-table-column label="状态" width="120"><template #default="{ row }"><StatusBadge type="product" :value="row.status" /></template></el-table-column>
            <el-table-column label="浏览" width="100" prop="view_count" />
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link type="primary" @click="$router.push(`/products/${row.id}`)">查看</el-button>
                <el-button v-if="row.status === 'on_sale'" link type="warning" @click="offShelf(row.id)">下架</el-button>
              </template>
            </el-table-column>
          </el-table>
          <EmptyState v-else description="还没有发布过商品" action-text="去发布" @action="$router.push('/products/create')" />
        </div>
      </el-tab-pane>
      <el-tab-pane label="我的收藏" name="favorites">
        <div v-loading="subLoading" class="fav-grid">
          <ProductCard v-for="p in favorites" :key="p.id" :product="p" />
        </div>
        <EmptyState v-if="!subLoading && !favorites.length" description="还没有收藏" />
      </el-tab-pane>
      <el-tab-pane label="收货地址" name="addresses">
        <el-table :data="addresses" v-loading="subLoading">
          <el-table-column prop="receiver_name" label="收货人" width="120" />
          <el-table-column prop="phone" label="手机号" width="140" />
          <el-table-column label="地址">
            <template #default="{ row }">{{ row.province }}{{ row.city }}{{ row.district }}{{ row.detail }}</template>
          </el-table-column>
          <el-table-column label="默认" width="80">
            <template #default="{ row }"><el-tag v-if="row.is_default" size="small" type="success">默认</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button link type="danger" @click="removeAddress(row.id)">删除</el-button>
              <el-button link type="primary" @click="addAddress">新增</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </el-card>

  <el-dialog v-model="profileDialog" title="编辑资料" width="420px">
    <el-form :model="profileForm" label-width="70px">
      <el-form-item label="昵称"><el-input v-model="profileForm.nickname" /></el-form-item>
      <el-form-item label="邮箱"><el-input v-model="profileForm.email" /></el-form-item>
      <el-form-item label="手机号"><el-input v-model="profileForm.phone" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="profileDialog = false">取消</el-button>
      <el-button type="primary" @click="saveProfile">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="addrDialog" title="新增收货地址" width="480px">
    <el-form :model="addrForm" label-width="80px">
      <el-form-item label="收货人"><el-input v-model="addrForm.receiver_name" /></el-form-item>
      <el-form-item label="手机号"><el-input v-model="addrForm.phone" /></el-form-item>
      <el-form-item label="省份"><el-input v-model="addrForm.province" /></el-form-item>
      <el-form-item label="城市"><el-input v-model="addrForm.city" /></el-form-item>
      <el-form-item label="区县"><el-input v-model="addrForm.district" /></el-form-item>
      <el-form-item label="详细地址"><el-input v-model="addrForm.detail" /></el-form-item>
      <el-form-item label="设为默认"><el-switch v-model="addrForm.is_default" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="addrDialog = false">取消</el-button>
      <el-button type="primary" @click="saveAddress">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import * as addressApi from '../api/address'
import { useUserStore } from '../stores/userStore'
import ProductCard from '../components/ProductCard.vue'
import StatusBadge from '../components/StatusBadge.vue'
import EmptyState from '../components/EmptyState.vue'
import { formatPrice, formatRole } from '../utils/format'

const userStore = useUserStore()
const tab = ref('products')
const loading = ref(false)
const subLoading = ref(false)
const myProducts = ref<any[]>([])
const favorites = ref<any[]>([])
const addresses = ref<any[]>([])
const profileDialog = ref(false)
const profileForm = ref({ nickname: '', email: '', phone: '' })
const addrDialog = ref(false)
const addrForm = ref({ receiver_name: '', phone: '', province: '', city: '', district: '', detail: '', is_default: false })

onMounted(async () => {
  loading.value = true
  try {
    if (!userStore.user) await userStore.fetchProfile()
    await Promise.all([loadProducts(), loadFavorites(), loadAddresses()])
  } finally {
    loading.value = false
  }
})

async function loadProducts() {
  subLoading.value = true
  try {
    const res: any = await productApi.myProducts({ page: 1, page_size: 50 })
    myProducts.value = res.data.list || []
  } finally {
    subLoading.value = false
  }
}

async function loadFavorites() {
  const res: any = await productApi.myFavorites({ page: 1, page_size: 50 })
  favorites.value = res.data.list || []
}

async function loadAddresses() {
  const res: any = await addressApi.listAddresses()
  addresses.value = res.data || []
}

async function offShelf(id: number) {
  await productApi.offShelfProduct(id)
  ElMessage.success('已下架')
  loadProducts()
}

function editProfile() {
  profileForm.value = { nickname: userStore.user?.nickname || '', email: userStore.user?.email || '', phone: userStore.user?.phone || '' }
  profileDialog.value = true
}

async function saveProfile() {
  await userStore.updateProfile(profileForm.value)
  profileDialog.value = false
  ElMessage.success('资料已更新')
}

function addAddress() {
  addrForm.value = { receiver_name: '', phone: '', province: '', city: '', district: '', detail: '', is_default: false }
  addrDialog.value = true
}

async function saveAddress() {
  await addressApi.createAddress(addrForm.value)
  addrDialog.value = false
  ElMessage.success('地址已保存')
  loadAddresses()
}

async function removeAddress(id: number) {
  await addressApi.deleteAddress(id)
  ElMessage.success('地址已删除')
  loadAddresses()
}
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.credit {
  color: #909399;
}
.edit-btn {
  margin-left: auto;
}
.fav-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
</style>
