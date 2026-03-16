<script setup lang="ts">
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()

const titleMap: Record<string, string> = {
  dashboard: '仪表盘',
  roles: '角色管理',
  users: '用户管理',
  customers: '客户管理',
  pool: '公海客户',
  search: '查找客户',
  potential: '潜在客户',
  partner: '合作客户',
  notifications: '通知中心',
  settings: '系统设置',
  profile: '个人资料',
}

const breadcrumbSegments = computed(() => {
  const segments = route.path.split('/').filter(Boolean)
  if (!segments.length) return [{ title: '仪表盘', url: '/' }]
  return segments.reduce<Array<{ title: string; url: string }>>((items, seg, i) => {
    const path = '/' + segments.slice(0, i + 1).join('/')
    const item = {
      title: titleMap[seg] || (route.meta?.title as string) || seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, ' '),
      url: path,
    }

    const previous = items[items.length - 1]
    if (previous?.title === item.title) {
      items[items.length - 1] = item
      return items
    }

    items.push(item)
    return items
  }, [])
})
</script>

<template>
  <Breadcrumb>
    <BreadcrumbList>
      <template v-for="(segment, index) in breadcrumbSegments" :key="index">
        <template v-if="index < breadcrumbSegments.length - 1">
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink as-child>
              <RouterLink :to="segment.url!">{{ segment.title }}</RouterLink>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
        </template>
        <BreadcrumbItem v-else class="hidden md:block">
          <BreadcrumbPage>{{ segment.title }}</BreadcrumbPage>
        </BreadcrumbItem>
      </template>
    </BreadcrumbList>
  </Breadcrumb>
</template>
