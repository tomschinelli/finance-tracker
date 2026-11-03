<template>
  <div class="app">
    <nav class="top-nav">
      <NuxtLink to="/" exact-active-class="active">Overview</NuxtLink>
      <NuxtLink to="/yearly" active-class="active">Yearly</NuxtLink>
      <NuxtLink to="/accounts" active-class="active">Accounts</NuxtLink>
      <NuxtLink to="/recurring" active-class="active">Recurring</NuxtLink>
    </nav>
    <main class="content">
      <NuxtPage />
    </main>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: system-ui, sans-serif;
  background: #f5f5f5;
  color: #222;
}

.top-nav {
  display: flex;
  gap: 1.5rem;
  padding: 1rem 1.5rem;
  background: #1f2937;
}

.top-nav a {
  color: #d1d5db;
  text-decoration: none;
  font-weight: 500;
}

.top-nav a.active,
.top-nav a:hover {
  color: #fff;
}

.content {
  max-width: 900px;
  margin: 0 auto;
  padding: 1.5rem;
}
</style>
