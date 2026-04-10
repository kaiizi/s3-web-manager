import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/buckets',
      name: 'BucketList',
      component: () => import('../views/BucketListView.vue'),
    },
    {
      path: '/buckets/:name',
      name: 'BucketDetail',
      component: () => import('../views/BucketDetailView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/buckets',
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.token) {
    return { name: 'Login' }
  }
})

export default router
