<script lang="ts">
import {api} from "../tool/api.ts";
import {MsgError, MsgSuccess, NewLoadingContext, ProcessResult} from "../tool/tool1.ts";
import type {SearchResult} from "../model/model.ts";
import C_info from "./info.vue"

export default {
  components: {
    C_info
  },
  data() {
    return {
      searchText: "",
      searchList: [] as SearchResult[],

      sourceList: [] as string[],
      sourceSelect: "",

      showImageIs: false,
      showImageData: "",
      showImageName: "",


      getInfoId: "",
      getInfoIs: false,

      lc: NewLoadingContext(),
    }
  },
  methods: {
    search() {
      if (this.searchText.length == 0) {
        MsgError("Please input search text")
        return
      }
      if (this.lc.isLoading) {
        MsgError("Previous search is still in progress...");
        return;
      }
      this.lc.Loading(async () => {
        const result = await api.Search(this.sourceSelect, this.searchText)
        const res = ProcessResult<SearchResult[]>(result)
        if (res) {
          this.searchList = res
          MsgSuccess(`Found ${this.searchList.length} results`)
        } else {
          MsgError("Search failed")
        }
      })
    },
    getSourceList() {
      this.lc.Loading(async () => {
        this.sourceList = await api.GetSourceList()
        if (this.sourceList.length > 0) {
          this.sourceSelect = this.sourceList[0]
        }
      })
    },
    getBase64Image(base64Data: string) {
      if (!base64Data.startsWith('data:image/')) {
        return `data:image/*;base64,${base64Data}`;
      }
      return base64Data;
    },
    showImage(name: string, base64Data: string) {
      this.showImageName = name
      this.showImageData = base64Data
      this.showImageIs = true
    },
    showInfo(id: string) {
      this.getInfoId = id
      this.getInfoIs = true
    },
  },
  mounted() {
    this.getSourceList()
  },
}
</script>

<template>
  <div v-loading="lc.isLoading" class="backgroundX">
    <div class="header">
      <div style="display: inline-flex; justify-items: center; align-items: center; gap: 10px;">
        <h2>NovelPackager</h2>
        <a target="_blank" rel="noopener noreferrer" href="https://github.com/peakedshout/novelpackager">
          <svg width="30px" height="30px" viewBox="0 0 97.707 96" xmlns="http://www.w3.org/2000/svg">
            <path fill-rule="evenodd" clip-rule="evenodd"
                  d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z"
                  fill="#24292f"/>
          </svg>
        </a>
      </div>
      <div class="search">
        <div style="display: inline-flex; justify-content: space-between; align-content: center;width: 100%">
          <el-input placeholder="Search..." v-model="searchText" style="flex: 1;" @keyup.enter="search"/>
          <el-button type="info" @click="search">🔍</el-button>
        </div>
      </div>
      <el-select style="width:10%; margin-right: 5%" v-model="sourceSelect" placeholder="Source">
        <el-option
            v-for="item in sourceList"
            :key="item"
            :label="item"
            :value="item"
        />
      </el-select>
    </div>
    <div v-if="searchList.length == 0">
      <el-empty description="No results found" style="margin: 0 auto;"></el-empty>
    </div>
    <div v-for="item in searchList">
      <div class="searchItem">
        <img :src="getBase64Image(item.cover)" alt="" style="width: auto; height: 100px;"
             @click="showImage(item.name,item.cover)">
        <div style="margin-left: 10px; text-align: left; width: 100%; height: 100px">
          <div style="display: inline-flex; justify-content: space-between;width: 100%">
            <el-button type="text" @click="showInfo(item.id)">
              <el-text size="large" line-clamp="1" class="custom-title"> {{ item.name }}</el-text>
            </el-button>
            <div style="display: inline-flex;">
              <el-tag v-for="t in item.metas">{{ t }}</el-tag>
            </div>
          </div>
          <br/>
          <el-text size="default" line-clamp="1">{{ item.author }}</el-text>
          <br/>
          <el-text size="small" line-clamp="2">{{ item.description }}</el-text>
        </div>
      </div>
    </div>
    <el-dialog v-model="showImageIs" :title="showImageName" width="50%"
               :before-close="() => {showImageIs = false;showImageName='';}">
      <img :src="getBase64Image(showImageData)" alt="" style="width: 100%; height: 100%">
    </el-dialog>
  </div>
  <C_info v-model:get-info-is="getInfoIs" v-model:get-info-id="getInfoId"
          v-model:source-select="sourceSelect"/>
</template>

<style scoped>
.custom-title:hover {
  color: deepskyblue;
}

.searchItem {
  width: 100%;
  height: 140px;
  margin: 0 auto;
  color: black;
  background-color: whitesmoke;

  display: inline-flex;
  justify-content: flex-start;
  align-items: center;

  border-radius: 2px;
  padding: 0 20px;
  box-sizing: border-box;
  border: 1px solid dimgray;
}

.search {
  width: 50%;
  border: 1px solid #ccc;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  color: black;
  background-color: darkgray;
  align-content: center;

  box-sizing: border-box;
}

.header {
  display: flex;
  background-color: dimgray;
  width: 100%;
  height: 80px;
  justify-content: space-between;
  align-items: center;
  border-radius: 2px;

  padding: 0 20px;
  box-sizing: border-box;
}

.backgroundX {
  background-color: white;
  border-radius: 2px;
  width: 100%;
  height: 100%;
}

</style>