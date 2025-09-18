import { createRouter, createWebHistory } from 'vue-router'

// 懒加载页面组件
// @ts-ignore
const LoginPage = () => import('../views/LoginPage.vue')
// @ts-ignore
const RegisterPage = () => import('../views/RegisterPage.vue')
// @ts-ignore
const DriveMain = () => import('../views/DriveMain.vue')
// @ts-ignore
const MountStorage = () => import('../views/MountStorage.vue')

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/login' // 默认重定向到登录页
    },
    {
      path: '/login',
      name: 'login',
      component: LoginPage,
      meta: {
        title: '登录',
        requiresAuth: false // 不需要登录即可访问
      }
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterPage,
      meta: {
        title: '注册',
        requiresAuth: false // 不需要登录即可访问
      }
    },
    {
      path: '/home',
      name: 'home',
      component: DriveMain,
      meta: {
        title: '网盘主页',
  requiresAuth: true // 需要登录才能访问
      }
    },
    {
      path: '/mount',
      name: 'mount',
      component: MountStorage,
      meta: {
        title: '挂载存储',
        requiresAuth: true // 需要登录才能访问
      }
    }
  ],
  // 路由切换时滚动到顶部
  scrollBehavior() {
    return { top: 0 };
  }
});

// 路由守卫：设置页面标题和登录验证
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title}`;
  }

  // 开发调试时取消鉴权强制跳转，直接继续路由。
  // 如果需要恢复鉴权，请把下面这行替换为原有的鉴权逻辑。
  next();
});

export default router

