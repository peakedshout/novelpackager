<script lang="ts">

import {api} from "./tool/api.ts";
import C_home from "./components/home.vue"

export default {
  components: {
    C_home
  },
  data() {
    return {
      searchText: "",
      sourceList: [] as string[],
    }
  },
  methods: {
    async getSourceList() {
      this.sourceList = await api.GetSourceList()
    }
  },
  mounted() {
    this.getSourceList()
  },
}
</script>

<template>
  <C_home/>
</template>

<style scoped>


</style>
