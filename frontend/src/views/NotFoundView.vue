<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面不存在</p>
    <router-link to="/">返回首页</router-link>
  </div>
</template>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  gap: 16px;
}
.not-found h1 {
  font-size: 72px;
  margin: 0;
  color: #409eff;
}
.not-found p {
  font-size: 18px;
  color: #909399;
}
</style>
