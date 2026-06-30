const plugins = [];

try {
  require.resolve('babel-plugin-module-resolver');
  plugins.push([
    'module-resolver',
    {
      root: ['./src'],
      alias: {
        '@': './src',
      },
      extensions: ['.ios.ts', '.android.ts', '.ts', '.tsx', '.js', '.json'],
    },
  ]);
} catch (error) {
  // Local node_modules may not be fully installed yet.
}

try {
  require.resolve('react-native-worklets/plugin');
  plugins.push('react-native-reanimated/plugin');
} catch (error) {
  // Reanimated plugin is enabled automatically once worklets is installed.
}

module.exports = {
  presets: ['module:@react-native/babel-preset', 'nativewind/babel'],
  plugins,
};
