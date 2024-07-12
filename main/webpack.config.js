const path = require("path")

module.exports = {
  entry: {
    bundle: "./static/js/index.js"
  },

  output: {
    filename: "[name].js",
    path: path.resolve(__dirname, "./static/js/bundle")
  },

  mode: "production",
  devtool: "source-map",

  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: [
          /node_modules/
        ],
        use: [
          { loader: "babel-loader" }
        ]
      }
    ]
  }
}
