import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {   // 首页
            path: '/',
            name: 'home',
            component: () => import('../views/HomeView.vue'),
        },
        {   // 账户（登录/注册合一）
            path: '/account',
            name: 'account',
            component: () => import('../views/AccountView.vue'),
        },
        {   // 设置
            path: '/settings',
            name: 'settings',
            meta: { requiresAuth: true },  // 需要登录
            component: () => import('../views/SettingsView.vue'),
        },
        {   // 视频详情
            path: '/video/:id',
            name: 'video-detail',
            component: () => import('../views/VideoDetailView.vue'),
        },
        {   // 热榜
            path: '/hot',
            name: 'hot',
            component: () => import('../views/HotView.vue'),
        },
        {   // 发布视频
            path: '/video/publish',
            name: 'publish',
            meta: { requiresAuth: true },  // 需要登录
            component: () => import('../views/PublishView.vue'),
        },
        {   // 我的（个人资料）
            path: '/profile',
            name: 'profile',
            meta: { requiresAuth: true },  // 需要登录
            component: () => import('../views/ProfileView.vue'),
        },
        {   // 404 兜底路由
            path: '/:pathMatch(.*)*',
            name: 'not-found',
            component: () => import('../views/NotFoundView.vue'),
        },
    ],
})

// 路由守卫：需要登录的页面未登录时跳转到登录页
router.beforeEach((to, _from, next) => {
    const auth = useAuthStore()   // 获取登录状态
    if (to.meta.requiresAuth && !auth.isLoggedIn) {
        next({ path: '/account', query: { redirect: to.fullPath } })
        return
    }
    next()
})

export default router
