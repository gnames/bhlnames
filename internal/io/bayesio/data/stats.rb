#!/usr/bin/env ruby
# frozen_string_literal: true

require 'json'

txt = File.read('bayes.json')
data = JSON.parse(txt, symbolize_names: true)
nomen = data[:labelCases][:isNomen].to_f
not_nomen = data[:labelCases][:notNomen].to_f

features = %i[annot title vol yrPage bestRes resNum]

features.each do |f|
  data[:featureCases][f].each do |k, v|
    likelihood = (v[:isNomen].to_f / nomen) / (v[:notNomen].to_f / not_nomen)
    puts format("%s:%s,%0.3f", f, k, likelihood)
  end
end
