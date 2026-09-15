// usePagination.ts 分页 hook：列表页统一复用。
import { ref } from 'vue'

export function usePagination(defaultSize = 10) {
  const page = ref(1)
  const pageSize = ref(defaultSize)
  const total = ref(0)

  function pageParams() {
    return { page: page.value, page_size: pageSize.value }
  }

  function setPage(p: number) {
    page.value = p
  }

  return { page, pageSize, total, pageParams, setPage }
}
