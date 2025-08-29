<template>
  <div class="login-card">
    <div class="logo-and-slogan">
      <h1 class="project-logo">项目logo</h1>
      <p class="slogan">一句项目广告词，或是随机寄语</p>
    </div>
    <form @submit.prevent="handleLogin" class="login-form">
      <div class="form-group">
        <label for="username">用户名:</label>
        <input v-model="username" type="text" id="username" placeholder="请输入用户名" required />
      </div>
      <div class="form-group">
        <label for="password">密码:</label>
        <input v-model="password" type="password" id="password" placeholder="请输入密码" required />
      </div>
      <div class="button-group">
        <button type="submit" class="login-button">登录</button>
        <button type="button" @click="handleRegister" class="register-button">注册</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')

const handleLogin = () => {
  // 这里可以添加登录逻辑，比如调用接口验证用户名密码
  ElMessage.success('登录成功，这里可以跳转到首页等操作')
  // 示例：router.push('/home')
}

const handleRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.login-card {
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  max-width: 400px;
  width: 100%;
  margin: 20px;
  text-align: center;
}

.logo-and-slogan {
  margin-bottom: 30px;
}

.project-logo {
  font-size: 32px;
  font-weight: bold;
  color: #fff;
  margin-bottom: 10px;
}

.slogan {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.8);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: left;
}

.form-group label {
  font-weight: 500;
  color: #fff;
  font-size: 14px;
}

.form-group input {
  padding: 12px 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 10px;
  font-size: 16px;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.form-group input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.button-group {
  display: flex;
  gap: 15px;
  margin-top: 10px;
}

.login-button,
.register-button {
  flex: 1;
  color: white;
  border: none;
  padding: 14px 20px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
}

.login-button {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
}

.login-button:active:not(:disabled) {
  transform: translateY(0);
}

.register-button {
  background: linear-gradient(135deg, #1f4ce1 0%, #7942d1 100%);
}

.register-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(108, 117, 125, 0.3);
}

.register-button:active {
  transform: translateY(0);
}

@media (max-width: 480px) {
  .login-card {
    padding: 30px 20px;
    margin: 10px;
  }
  .button-group {
    flex-direction: column;
    gap: 10px;
  }
  .login-button,
  .register-button {
    width: 100%;
  }
}
</style>
