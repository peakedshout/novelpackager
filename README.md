# novelpackager
### (EN/[CN](./README_CN.md))
A novel crawler packaging epub tool implemented in pure golang using the rod framework. / 使用rod框架，纯golang实现的小说爬虫打包epub工具。

## Description
- This tool uses the rod framework to crawl novel websites, obtain novel information, download novels, and package epub.
  - If you don't know what rod is? [Click here](https://github.com/go-rod/rod)
- Essentially, the browser will be used to perform operations, and the operations performed are:
  - Retrieve an executable browser -> If not, retrieve from the parameters -> If still not, download one according to the operating system -> Then execute the automation process according to the source.
  - Generally speaking, if the execution environment has a browser, there will be basically no problem.
- Currently there are not many supported sources, so just make do with what you have.
- Want to use a different source?
  - You can import your source in the pkg/boot directory. The specific implementation should be something you should consider.
  - In the future, you may consider a more flexible import method, but this should be considered after development to a certain extent. I don’t have a good idea at the moment.
- Why use rod?
  - rod is very useful. For crawlers, nothing is more convenient than operating on a real browser.
  - Note that for rod, the browser is used to automatically execute the logic, but as we all know, the browser itself will take up a lot of resources, so make sure the running environment is sufficient.
  - In the future, we may consider using rod's management mode to separate: the expansion mode of the commander and the executive (so that it can also support the acquisition of data on mobile phones (or niche environments) in the future, but the specific logic is executed in an environment with more sufficient resources, which is very attractive.
- Can the server's operating system be used?
  - The answer is yes, because this project is completely based on golang, so in theory any operating system that supports golang can run it.
  - And for non-graphical pages, the browser has a headless mode and can also run on a pure terminal. (Tested on Ubuntu, you can see the pkg/note file to install some dependencies (essentially browser dependencies))
  - So all you need to do is make sure the environment can execute the browser.
- If you have any questions about this project, please raise an issue and I will try to respond in a timely manner.

## Usage
- There is no executable command currently, see the command implemented in the specific source.

## Source list (click the link to view detailed source description)
- [x] [bilinovel](./pkg/source/bilinovel) bilinovel is a light novel source with frequent updates and many anti-crawler strategies, but the automation attributes of rod can also be well adapted.
- [ ] other

## TODO
- [x] webui (Use `novelpackager web` to start, and then operate through webui. You can see the parameters for specific settings. Simple cache optimization is built in)
- [ ] Add timed check update logic (used to obtain updated chapters or volumes in time)
- [ ] Remote operation mode
- [x] Currently it has satisfied my personal use (downloaded offline content and reading it 😊)
- [ ] Support comic packaging...?
- [ ] More sources...
- [ ] Others...