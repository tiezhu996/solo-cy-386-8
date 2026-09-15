<template>
  <el-card v-loading="loading">
    <h2>确认订单</h2>
    <template v-if="product">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="商品">{{ product.title }}</el-descriptions-item>
        <el-descriptions-item label="卖家">{{ product.seller?.nickname || `用户${product.seller_id}` }}</el-descriptions-item>
        <el-descriptions-item label="单价">¥{{ formatPrice(product.price) }}</el-descriptions-item>
        <el-descriptions-item label="数量">{{ quantity }}</el-descriptions-item>
        <el-descriptions-item label="合计"><b class="total">¥{{ formatPrice(product.price * quantity) }}</b></el-descriptions-item>
      </el-descriptions>

      <div class="block">
        <h3>选择收货地址</h3>
        <el-radio-group v-model="addressId" v-if="addresses.length">
          <el-radio v-for="a in addresses" :key="a.id" :value="a.id" class="addr-radio">
            {{ a.receiver_name }} {{ a.phone }} · {{ a.province }}{{ a.city }}{{ a.district }}{{ a.detail }}
            <el-tag v-if="a.is_default" size="small" type="success">默认</el-tag>
          </el-radio>
        </el-radio-group>
        <EmptyState v-else description="还没有收货地址，请先添加" />
        <div class="addr-actions">
          <el-button size="small" @click="showAddrDialog = true">新增地址</el-button>
          <router-link to="/profile"><el-button size="small" text>管理地址</el-button></router-link>
        </div>
      </div>

      <el-form-item label="备注">
        <el-input v-model="remark" placeholder="给卖家的留言（选填）" />
      </el-form-item>

      <el-button type="danger" size="large" :loading="submitting" :disabled="!addressId" @click="submit">提交订单</el-button>
    </template>
  </el-card>

  <el-dialog v-model="showAddrDialog" title="新增收货地址" width="480px">
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
      <el-button @click="showAddrDialog = false">取消</el-button>
      <el-button type="primary" @click="saveAddress">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import * as addressApi from '../api/address'
import * as orderApi from '../api/order'
import EmptyState from '../components/EmptyState.vue'
import { formatPrice } from '../utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const product = ref<any>(null)
const addresses = ref<any[]>([])
const addressId = ref<number | null>(null)
const quantity = 1
const remark = ref('')
const showAddrDialog = ref(false)
const addrForm = ref({ receiver_name: '', phone: '', province: '', city: '', district: '', detail: '', is_default: false })

onMounted(async () => {
  loading.value = true
  try {
    const pid = Number(route.query.product_id)
    if (!pid) {
      ElMessage.warning('缺少商品参数')
      router.push('/')
      return
    }
    const [prodRes, addrRes]: any = await Promise.all([productApi.getProduct(pid), addressApi.listAddresses()])
    product.value = prodRes.data
    addresses.value = addrRes.data || []
    const def = addresses.value.find((a) => a.is_default)
    addressId.value = def ? def.id : (addresses.value[0]?.id ?? null)
  } finally {
    loading.value = false
  }
})

async function saveAddress() {
  const res: any = await addressApi.createAddress(addrForm.value)
  addresses.value.push(res.data)
  if (!addressId.value) addressId.value = res.data.id
  showAddrDialog.value = false
  ElMessage.success('地址已保存')
}

async function submit() {
  if (!addressId.value) {
    ElMessage.warning('请选择收货地址')
    return
  }
  submitting.value = true
  try {
    const res: any = await orderApi.createOrder({ product_id: product.value.id, address_id: addressId.value, quantity, remark: remark.value })
    ElMessage.success('下单成功')
    router.push('/orders')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.block {
  margin: 20px 0;
}
.addr-radio {
  display: block;
  margin: 8px 0;
}
.addr-actions {
  margin-top: 8px;
}
.total {
  color: #f56c6c;
  font-size: 18px;
}
</style>
